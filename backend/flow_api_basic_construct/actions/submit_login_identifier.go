package actions

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"strings"
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
		c.AddInputs(
			flowpilot.EmailInput("identifier").
				Required(true).
				Preserve(true).
				MaxLength(255))
	} else {
		c.AddInputs(
			flowpilot.StringInput("identifier").
				Required(true).
				Preserve(true).
				MinLength(a.cfg.Identifier.Username.MinLength).
				MaxLength(a.cfg.Identifier.Username.MaxLength))
	}

	// TODO: suspend action when no other login method other than oauth is available
}

func (a SubmitLoginIdentifier) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	identifier := c.Input().Get("identifier").String()

	// TODO: Maybe think of better check?
	if strings.Contains(identifier, "@") {
		// User has submitted an email address.

		if err := c.Stash().Set("email", identifier); err != nil {
			return fmt.Errorf("failed to set email to stash: %w", err)
		}

		if a.cfg.Password.Enabled {
			if a.cfg.Password.Optional {
				return c.ContinueFlow(common.StateLoginMethodChooser)
			} else {
				return c.ContinueFlow(common.StatePasswordLogin)
			}
		}

		return c.ContinueFlow(common.StateLoginPasscodeConfirmation)
	}

	if a.cfg.Password.Enabled {
		if a.cfg.Password.Optional {
			return c.ContinueFlow(common.StateLoginMethodChooser)
		} else {
			return c.ContinueFlow(common.StatePasswordLogin)
		}
	}

	username, err := a.persister.GetUsernamePersister().Find(identifier)
	if err != nil {
		return err
	}

	if username == nil {
		c.Input().SetError("identifier", flowpilot.ErrorValueInvalid.Wrap(errors.New("username not found")))
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	if err = c.Stash().Set("username", identifier); err != nil {
		return fmt.Errorf("failed to set username to stash: %w", err)
	}

	user, err := a.persister.GetUserPersister().Get(username.UserID)
	if err != nil {
		return err
	}

	if primaryEmail := user.Emails.GetPrimary(); primaryEmail != nil {
		if err = c.Stash().Set("email", primaryEmail.Address); err != nil {
			return fmt.Errorf("failed to set email to stash: %w", err)
		}

		return c.ContinueFlow(common.StateLoginPasscodeConfirmation)
	}

	// Username exists, but user has no emails.
	return c.ContinueFlow(common.StateLoginMethodChooser)
}
