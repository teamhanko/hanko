package actions

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"regexp"
)

func NewSubmitLoginIdentifier(cfg config.Config, persister persistence.Persister, httpContext echo.Context) SubmitLoginIdentifier {
	return SubmitLoginIdentifier{
		cfg,
		persister,
		httpContext,
	}
}

type SubmitLoginIdentifier struct {
	cfg         config.Config
	persister   persistence.Persister
	httpContext echo.Context
}

func (a SubmitLoginIdentifier) GetName() flowpilot.ActionName {
	return common.ActionSubmitLoginIdentifier
}

func (a SubmitLoginIdentifier) GetDescription() string {
	return "Enter an identifier to login."
}

func (a SubmitLoginIdentifier) Initialize(c flowpilot.InitializationContext) {
	if !a.cfg.Identifier.Username.Enabled {
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

	if !a.cfg.Password.Enabled && !a.cfg.Passcode.Enabled {
		c.SuspendAction()
	}
}

func (a SubmitLoginIdentifier) Execute(c flowpilot.ExecutionContext) error {
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
		userModel, err := a.persister.GetUserPersister().GetByUsername(identifier)
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

	if a.cfg.Password.Enabled {
		if a.cfg.Password.Optional {
			return c.ContinueFlow(common.StateLoginMethodChooser)
		} else {
			return c.ContinueFlow(common.StateLoginPassword)
		}
	}

	if c.Stash().Get("email").Exists() {
		if err := c.Stash().Set("passcode_template", "login"); err != nil {
			return fmt.Errorf("failed to set passcode_template to stash: %w", err)
		}
		return c.ContinueFlow(common.StateLoginPasscodeConfirmation)
	}

	// Username exists, but user has no emails.
	return c.ContinueFlow(common.StateLoginMethodChooser)
}
