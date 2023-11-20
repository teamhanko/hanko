package login

import (
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"regexp"
)

type SubmitLoginIdentifier struct {
	shared.Action
}

func (a SubmitLoginIdentifier) GetName() flowpilot.ActionName {
	return ActionSubmitLoginIdentifier
}

func (a SubmitLoginIdentifier) GetDescription() string {
	return "Enter an identifier to login."
}

func (a SubmitLoginIdentifier) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Identifier.Username.Enabled {
		input := flowpilot.EmailInput("identifier").
			Required(true).
			Preserve(true).
			MaxLength(255)

		c.AddInputs(input)
	} else {
		input := flowpilot.StringInput("identifier").
			Required(true).
			Preserve(true).
			MaxLength(255)

		c.AddInputs(input)
	}

	if !deps.Cfg.Password.Enabled && !deps.Cfg.Passcode.Enabled {
		c.SuspendAction()
	}
}

func (a SubmitLoginIdentifier) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	identifier := c.Input().Get("identifier").String()
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	if isEmail := emailPattern.MatchString(identifier); isEmail {
		// User has submitted an email address.

		if err := c.Stash().Set("email", identifier); err != nil {
			return fmt.Errorf("failed to set email to stash: %w", err)
		}
	} else {
		userModel, err := deps.Persister.GetUserPersister().GetByUsername(identifier)
		if err != nil {
			return err
		}

		if userModel == nil {
			c.Input().SetError("identifier", flowpilot.ErrorValueInvalid.Wrap(errors.New("username not found")))
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		}

		if err = c.Stash().Set("username", identifier); err != nil {
			return fmt.Errorf("failed to set username to stash: %w", err)
		}

		if primaryEmailModel := userModel.Emails.GetPrimary(); primaryEmailModel != nil {
			if err = c.Stash().Set("email", primaryEmailModel.Address); err != nil {
				return fmt.Errorf("failed to set email to stash: %w", err)
			}
		}
	}

	if deps.Cfg.Password.Enabled {
		if deps.Cfg.Password.Optional {
			return c.ContinueFlow(StateLoginMethodChooser)
		} else {
			return c.ContinueFlow(StateLoginPassword)
		}
	}

	if c.Stash().Get("email").Exists() {
		if err := c.Stash().Set("passcode_template", "login"); err != nil {
			return fmt.Errorf("failed to set passcode_template to stash: %w", err)
		}

		return c.StartSubFlow(passcode.StatePasscodeConfirmation, shared.StateSuccess)
	}

	// Username exists, but user has no emails.
	return c.ContinueFlow(StateLoginMethodChooser)
}
