package login

import (
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"regexp"
)

type ContinueWithLoginIdentifier struct {
	shared.Action
}

func (a ContinueWithLoginIdentifier) GetName() flowpilot.ActionName {
	return ActionContinueWithLoginIdentifier
}

func (a ContinueWithLoginIdentifier) GetDescription() string {
	return "Enter an identifier to login."
}

func (a ContinueWithLoginIdentifier) Initialize(c flowpilot.InitializationContext) {
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

func (a ContinueWithLoginIdentifier) Execute(c flowpilot.ExecutionContext) error {
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

		emailModel, err := deps.Persister.GetEmailPersister().FindByAddress(identifier)
		if err != nil {
			return fmt.Errorf("failed to get email model from db: %w", err)
		}

		if emailModel != nil && emailModel.UserID != nil {
			err := c.Stash().Set("user_id", emailModel.UserID.String())
			if err != nil {
				return fmt.Errorf("failed to set user_id to the stash: %w", err)
			}
		} else {
			err = c.Stash().Set("passcode_template", "login_attempted")
			if err != nil {
				return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
			}

			return c.StartSubFlow(passcode.StatePasscodeConfirmation)
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

		err = c.Stash().Set("user_id", userModel.ID.String())
		if err != nil {
			return fmt.Errorf("failed to set user_id to the stash: %w", err)
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

		if deps.Cfg.Passkey.Onboarding.Enabled && c.Stash().Get("webauthn_available").Bool() {
			return c.StartSubFlow(passcode.StatePasscodeConfirmation, passkey_onboarding.StateOnboardingCreatePasskey, shared.StateSuccess)
		} else {
			return c.StartSubFlow(passcode.StatePasscodeConfirmation, shared.StateSuccess)
		}
	}

	// Username exists, but user has no emails.
	return c.ContinueFlow(StateLoginMethodChooser)
}
