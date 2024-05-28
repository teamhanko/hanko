package login

import (
	"errors"
	"fmt"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"regexp"
)

type ContinueWithLoginIdentifier struct {
	shared.Action
}

func (a ContinueWithLoginIdentifier) GetName() flowpilot.ActionName {
	return shared.ActionContinueWithLoginIdentifier
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
		input = flowpilot.StringInput("identifier").
			MaxLength(255)
	} else if emailEnabled {
		input = flowpilot.EmailInput("email").
			MaxLength(deps.Cfg.Email.MaxLength).
			MinLength(3)
	} else if usernameEnabled {
		input = flowpilot.StringInput("username").
			MaxLength(deps.Cfg.Username.MaxLength).
			MinLength(deps.Cfg.Username.MinLength)
	}

	if input != nil {
		c.AddInputs(input.
			Required(true).
			Preserve(true))
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

		if userModel != nil {
			emailModel := userModel.GetEmailByAddress(identifierInputValue)

			if emailModel != nil && emailModel.UserID != nil {
				err := c.Stash().Set("user_id", emailModel.UserID.String())
				if err != nil {
					return fmt.Errorf("failed to set user_id to the stash: %w", err)
				}
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

	var onboardingStates []flowpilot.StateName
	if userModel != nil {
		var err error
		onboardingStates, err = a.determineOnboardingStates(c, userModel)
		if err != nil {
			return fmt.Errorf("failed to determine onboarding states: %w", err)
		}
	}

	if deps.Cfg.Email.UseForAuthentication && deps.Cfg.Password.Enabled {
		if treatIdentifierAsEmail || (!treatIdentifierAsEmail && userModel != nil && userModel.Emails.GetPrimary() != nil) {
			return c.StartSubFlow(shared.StateLoginMethodChooser, onboardingStates...)
		}

		return c.StartSubFlow(shared.StateLoginPassword, onboardingStates...)
	} else if deps.Cfg.Email.UseForAuthentication {
		// Set only for audit logging purposes.
		if err := c.Stash().Set("login_method", "passcode"); err != nil {
			return fmt.Errorf("failed to set login_method to stash: %w", err)
		}

		return c.StartSubFlow(shared.StatePasscodeConfirmation, onboardingStates...)
	} else if deps.Cfg.Password.Enabled {
		return c.StartSubFlow(shared.StateLoginPassword, onboardingStates...)
	}

	return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFlowDiscontinuity.Wrap(errors.New("no authentication method enabled")))
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

func (a ContinueWithLoginIdentifier) determineOnboardingStates(c flowpilot.ExecutionContext, userModel *models.User) ([]flowpilot.StateName, error) {
	deps := a.GetDeps(c)

	userHasPassword := userModel.PasswordCredential != nil
	userHasWebauthnCredential := len(userModel.WebauthnCredentials) > 0
	userHasUsername := len(userModel.Username) > 0
	userHasEmail := len(userModel.Emails) > 0

	if err := c.Stash().Set("user_has_password", userHasPassword); err != nil {
		return nil, fmt.Errorf("failed to set user_has_password to the stash: %w", err)
	}

	if err := c.Stash().Set("user_has_webauthn_credential", userHasWebauthnCredential); err != nil {
		return nil, fmt.Errorf("failed to set user_has_webauthn_credential to the stash: %w", err)
	}

	userDetailOnboardingStates := a.determineUserDetailOnboardingStates(deps.Cfg, userHasUsername, userHasEmail)
	credentialOnboardingStates := a.determineCredentialOnboardingStates(deps.Cfg, userHasWebauthnCredential, userHasPassword)

	return append(userDetailOnboardingStates, append(credentialOnboardingStates, shared.StateSuccess)...), nil
}

func (a ContinueWithLoginIdentifier) determineCredentialOnboardingStates(cfg config.Config, hasPasskey, hasPassword bool) []flowpilot.StateName {
	result := make([]flowpilot.StateName, 0)

	if cfg.Passkey.AcquireOnLogin == "always" && cfg.Password.AcquireOnLogin == "always" {
		if !hasPasskey && !hasPassword {
			result = append(result, shared.StateOnboardingCreatePasskey, shared.StatePasswordCreation)
		} else if hasPasskey && !hasPassword {
			result = append(result, shared.StatePasswordCreation)
		} else if !hasPasskey && hasPassword {
			result = append(result, shared.StateOnboardingCreatePasskey)
		}
	} else if cfg.Passkey.AcquireOnLogin == "always" && cfg.Password.AcquireOnLogin == "conditional" {
		if !hasPasskey && !hasPassword {
			result = append(result, shared.StateOnboardingCreatePasskey) // skip should lead to password onboarding
		} else if !hasPasskey && hasPassword {
			result = append(result, shared.StateOnboardingCreatePasskey)
		}
	} else if cfg.Passkey.AcquireOnLogin == "conditional" && cfg.Password.AcquireOnLogin == "always" {
		if !hasPasskey && !hasPassword {
			result = append(result, shared.StatePasswordCreation) // skip should lead to passkey onboarding
		} else if hasPasskey && !hasPassword {
			result = append(result, shared.StatePasswordCreation)
		}
	} else if cfg.Passkey.AcquireOnLogin == "conditional" && cfg.Password.AcquireOnLogin == "conditional" {
		if !hasPasskey && !hasPassword {
			result = append(result, shared.StateCredentialOnboardingChooser) // credential_onboarding_chooser can be skipped
		}
	} else if cfg.Passkey.AcquireOnLogin == "conditional" && cfg.Password.AcquireOnLogin == "never" {
		if !hasPasskey && !hasPassword {
			result = append(result, shared.StateOnboardingCreatePasskey)
		}
	} else if cfg.Passkey.AcquireOnLogin == "never" && cfg.Password.AcquireOnLogin == "conditional" {
		if !hasPasskey && !hasPassword {
			result = append(result, shared.StatePasswordCreation)
		}
	} else if cfg.Passkey.AcquireOnLogin == "never" && cfg.Password.AcquireOnLogin == "always" {
		if !hasPassword {
			result = append(result, shared.StatePasswordCreation)
		}
	} else if cfg.Passkey.AcquireOnLogin == "always" && cfg.Password.AcquireOnLogin == "never" {
		if !hasPasskey {
			result = append(result, shared.StateOnboardingCreatePasskey)
		}
	}

	return result
}

func (a ContinueWithLoginIdentifier) determineUserDetailOnboardingStates(cfg config.Config, userHasUsername, userHasEmail bool) []flowpilot.StateName {
	result := make([]flowpilot.StateName, 0)

	acquireUsername := !userHasUsername && cfg.Username.AcquireOnLogin
	acquireEmail := !userHasEmail && cfg.Email.AcquireOnLogin

	if acquireUsername && acquireEmail {
		result = append(result, shared.StateOnboardingUsername, shared.StateOnboardingEmail)
	} else if acquireUsername {
		result = append(result, shared.StateOnboardingUsername)
	} else if acquireEmail {
		result = append(result, shared.StateOnboardingEmail)
	}

	if acquireEmail && cfg.Email.RequireVerification {
		result = append(result, shared.StatePasscodeConfirmation)
	}

	return result
}
