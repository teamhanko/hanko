package login

import (
	"fmt"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
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

	var input flowpilot.Input
	if deps.Cfg.Identifier.Username.Enabled && deps.Cfg.Identifier.Email.Enabled {
		input = flowpilot.StringInput("identifier")
	} else if deps.Cfg.Identifier.Email.Enabled {
		input = flowpilot.EmailInput("email")
	} else if deps.Cfg.Identifier.Username.Enabled {
		input = flowpilot.StringInput("username")
	}

	if input != nil {
		c.AddInputs(input.
			Required(true).
			Preserve(true).
			MinLength(3).
			MaxLength(255))
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

	identifierInputName, identifierInputValue, treatIdentifierAsEmail := a.analyzeIdentifierInputs(c)

	if len(identifierInputValue) == 0 {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	if treatIdentifierAsEmail {
		// User has submitted an email address.

		if err := c.Stash().Set("email", identifierInputValue); err != nil {
			return fmt.Errorf("failed to set email to stash: %w", err)
		}

		emailModel, err := deps.Persister.GetEmailPersister().FindByAddress(identifierInputValue)
		if err != nil {
			return fmt.Errorf("failed to get email model from db: %w", err)
		}

		if emailModel != nil && emailModel.UserID != nil {
			err := c.Stash().Set("user_id", emailModel.UserID.String())
			if err != nil {
				return fmt.Errorf("failed to set user_id to the stash: %w", err)
			}
		} else {
			err = c.Stash().Set("passcode_template", "email_login_attempted")
			if err != nil {
				return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
			}

			return c.StartSubFlow(passcode.StatePasscodeConfirmation)
		}
	} else {
		userModel, err := deps.Persister.GetUserPersister().GetByUsername(identifierInputValue)
		if err != nil {
			return err
		}

		if userModel == nil {
			flowError := shared.ErrorUnknownUsername
			err = deps.AuditLogger.CreateWithConnection(
				deps.Tx,
				deps.HttpContext,
				models.AuditLogLoginFailure,
				nil,
				flowError,
				auditlog.Detail("flow_id", c.GetFlowID()))

			if err != nil {
				return fmt.Errorf("could not create audit log: %w", err)
			}

			c.Input().SetError(identifierInputName, flowError)
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		}

		if err = c.Stash().Set("username", identifierInputValue); err != nil {
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

		// Set only for audit logging purposes.
		if err := c.Stash().Set("login_method", "passcode"); err != nil {
			return fmt.Errorf("failed to set login_method to stash: %w", err)
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

func (a ContinueWithLoginIdentifier) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}

// analyzeIdentifierInputs determines if an input value has been provided for 'identifier', 'email', or 'username',
// according to the configuration. Also adds an input error to the expected input field, if the value is missing.
// Returns the related input field name, the provided value, and a flag, indicating if the value should be treated as
// an email (and not as a username).
func (a ContinueWithLoginIdentifier) analyzeIdentifierInputs(c flowpilot.ExecutionContext) (name string, value string, treatAsEmail bool) {
	deps := a.GetDeps(c)
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	if deps.Cfg.Identifier.Username.Enabled && deps.Cfg.Identifier.Email.Enabled {
		// analyze the 'identifier' input field
		name = "identifier"
		value = c.Input().Get(name).String()
		treatAsEmail = emailPattern.MatchString(value)
	} else if deps.Cfg.Identifier.Email.Enabled {
		// analyze the 'email' input field
		name = "email"
		value = c.Input().Get(name).String()
		treatAsEmail = true
	} else if deps.Cfg.Identifier.Username.Enabled {
		// analyze the 'username' input field
		name = "username"
		value = c.Input().Get(name).String()
		treatAsEmail = false
	}

	// If no value could not be determined, set an error for the missing input
	if len(value) == 0 && len(name) > 0 {
		c.Input().SetError(name, flowpilot.ErrorValueMissing)
	}

	return name, value, treatAsEmail
}
