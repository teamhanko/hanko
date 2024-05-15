package login

import (
	"errors"
	"fmt"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/constants"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"regexp"
)

type ContinueWithLoginIdentifier struct {
	shared.Action
}

func (a ContinueWithLoginIdentifier) GetName() flowpilot.ActionName {
	return constants.ActionContinueWithLoginIdentifier
}

func (a ContinueWithLoginIdentifier) GetDescription() string {
	return "Enter an identifier to login."
}

func (a ContinueWithLoginIdentifier) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	emailEnabled := deps.Cfg.Email.Enabled && deps.Cfg.Email.UseAsLoginIdentifier
	usernameEnabled := deps.Cfg.Username.Enabled && deps.Cfg.Username.UseAsLoginIdentifier

	var input flowpilot.Input
	if usernameEnabled && emailEnabled {
		input = flowpilot.StringInput("identifier")
	} else if emailEnabled {
		input = flowpilot.EmailInput("email")
	} else if usernameEnabled {
		input = flowpilot.StringInput("username")
	}

	if input != nil {
		c.AddInputs(input.
			Required(true).
			Preserve(true).
			MinLength(3).
			MaxLength(255))
	}

	if (!deps.Cfg.Password.Enabled && !deps.Cfg.Email.UseForAuthentication) || (!emailEnabled && !usernameEnabled) {
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

	var userModel *models.User

	if treatIdentifierAsEmail {
		// User has submitted an email address.

		var err error
		userModel, err = deps.Persister.GetUserPersister().GetByEmailAddress(identifierInputValue)
		if err != nil {
			return err
		}

		if err = c.Stash().Set("email", identifierInputValue); err != nil {
			return fmt.Errorf("failed to set email to stash: %w", err)
		}

		emailModel := userModel.GetEmailByAddress(identifierInputValue)

		if emailModel != nil && emailModel.UserID != nil {
			err := c.Stash().Set("user_id", emailModel.UserID.String())
			if err != nil {
				return fmt.Errorf("failed to set user_id to the stash: %w", err)
			}
		}
	} else {
		// User has submitted a username.

		var err error
		userModel, err = deps.Persister.GetUserPersister().GetByUsername(identifierInputValue)
		if err != nil {
			return err
		}

		if userModel == nil {
			flowInputError := shared.ErrorUnknownUsername
			err = deps.AuditLogger.CreateWithConnection(
				deps.Tx,
				deps.HttpContext,
				models.AuditLogLoginFailure,
				nil,
				flowInputError,
				auditlog.Detail("flow_id", c.GetFlowID()))

			if err != nil {
				return fmt.Errorf("could not create audit log: %w", err)
			}

			c.Input().SetError(identifierInputName, flowInputError)
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

	onboardingStates := a.determineCredentialOnboardingStates(deps.Cfg, len(userModel.WebauthnCredentials) > 0, userModel.PasswordCredential != nil)
	onboardingStates = append(onboardingStates, constants.StateSuccess)

	if deps.Cfg.Email.UseForAuthentication && deps.Cfg.Password.Enabled {
		return c.StartSubFlow(constants.StateLoginMethodChooser, onboardingStates...)
	} else if deps.Cfg.Email.UseForAuthentication {
		// Set only for audit logging purposes.
		if err := c.Stash().Set("login_method", "passcode"); err != nil {
			return fmt.Errorf("failed to set login_method to stash: %w", err)
		}

		return c.StartSubFlow(constants.StatePasscodeConfirmation, onboardingStates...)
	} else if deps.Cfg.Password.Enabled {
		return c.StartSubFlow(constants.StateLoginPassword, onboardingStates...)
	}

	return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFlowDiscontinuity.Wrap(errors.New("no authentication method enabled")))

	//if c.Stash().Get("email").Exists() {
	//
	//
	//
	//	if deps.Cfg.Passkey.AcquireOnLogin == "always" && c.Stash().Get("webauthn_available").Bool() {
	//		return c.StartSubFlow(passcode.StatePasscodeConfirmation, passkey_onboarding.StateOnboardingCreatePasskey, shared.StateSuccess)
	//	} else {
	//		return c.StartSubFlow(passcode.StatePasscodeConfirmation, shared.StateSuccess)
	//	}
	//}

	// Username exists, but user has no emails.
	// return c.ContinueFlow(StateLoginMethodChooser)
}

func (a ContinueWithLoginIdentifier) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}

// analyzeIdentifierInputs determines if an input value has been provided for 'identifier', 'email', or 'username',
// according to the configuration. Also adds an input error to the expected input field, if the value is missing.
// Returns the related input field name, the provided value, and a flag, indicating if the value should be treated as
// an email (and not as a username).
func (a ContinueWithLoginIdentifier) analyzeIdentifierInputs(c flowpilot.ExecutionContext) (name, value string, treatAsEmail bool) {
	deps := a.GetDeps(c)
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	emailEnabled := deps.Cfg.Email.Enabled && deps.Cfg.Email.UseAsLoginIdentifier
	usernameEnabled := deps.Cfg.Username.Enabled && deps.Cfg.Username.UseAsLoginIdentifier

	if emailEnabled && usernameEnabled {
		// analyze the 'identifier' input field
		name = "identifier"
		value = c.Input().Get(name).String()
		treatAsEmail = emailPattern.MatchString(value)
	} else if emailEnabled {
		// analyze the 'email' input field
		name = "email"
		value = c.Input().Get(name).String()
		treatAsEmail = true
	} else if usernameEnabled {
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

func (a ContinueWithLoginIdentifier) determineCredentialOnboardingStates(cfg config.Config, hasPasskey, hasPassword bool) []flowpilot.StateName {
	result := make([]flowpilot.StateName, 0)

	if cfg.Passkey.AcquireOnLogin == "always" && cfg.Password.AcquireOnLogin == "always" {
		if !hasPasskey && !hasPassword {
			result = append(result, constants.StateOnboardingCreatePasskey, constants.StatePasswordCreation)
		} else if hasPasskey && !hasPassword {
			result = append(result, constants.StatePasswordCreation)
		} else if !hasPasskey && hasPassword {
			result = append(result, constants.StateOnboardingCreatePasskey)
		}
	} else if cfg.Passkey.AcquireOnLogin == "always" && cfg.Password.AcquireOnLogin == "conditional" {
		if !hasPasskey && !hasPassword {
			result = append(result, constants.StateOnboardingCreatePasskeyConditional) // skip should lead to password onboarding
		} else if !hasPasskey && hasPassword {
			result = append(result, constants.StateOnboardingCreatePasskey)
		}
	} else if cfg.Passkey.AcquireOnLogin == "conditional" && cfg.Password.AcquireOnLogin == "always" {
		if !hasPasskey && !hasPassword {
			result = append(result, constants.StatePasswordCreationConditional) // skip should lead to passkey onboarding
		} else if hasPasskey && !hasPassword {
			result = append(result, constants.StatePasswordCreation)
		}
	} else if cfg.Passkey.AcquireOnLogin == "conditional" && cfg.Password.AcquireOnLogin == "conditional" {
		if !hasPasskey && !hasPassword {
			if cfg.Passkey.Optional && cfg.Password.Optional {
				result = append(result, "login_method_onboarding_chooser") // login_method_onboarding_chooser can be skipped
			} else if cfg.Passkey.Optional && !cfg.Password.Optional {
				result = append(result, constants.StatePasswordCreation, constants.StateOnboardingCreatePasskey) // passkey_onboarding can be skipped
			} else if !cfg.Passkey.Optional && cfg.Password.Optional {
				result = append(result, constants.StateOnboardingCreatePasskey, constants.StatePasswordCreation) // password_onboarding can be skipped
			} else if !cfg.Passkey.Optional && !cfg.Password.Optional {
				result = append(result, constants.StateOnboardingCreatePasskey, constants.StatePasswordCreation) // both states cannot be skipped
			}
		}
	} else if cfg.Passkey.AcquireOnLogin == "conditional" && cfg.Password.AcquireOnLogin == "never" {
		if !hasPasskey && !hasPassword {
			result = append(result, constants.StateOnboardingCreatePasskey)
		}
	} else if cfg.Passkey.AcquireOnLogin == "never" && cfg.Password.AcquireOnLogin == "conditional" {
		if !hasPasskey && !hasPassword {
			result = append(result, constants.StatePasswordCreation)
		}
	} else if cfg.Passkey.AcquireOnLogin == "never" && cfg.Password.AcquireOnLogin == "always" {
		if !hasPassword {
			result = append(result, constants.StatePasswordCreation)
		}
	} else if cfg.Passkey.AcquireOnLogin == "always" && cfg.Password.AcquireOnLogin == "never" {
		if !hasPasskey {
			result = append(result, constants.StateOnboardingCreatePasskey)
		}
	}

	return result
}
