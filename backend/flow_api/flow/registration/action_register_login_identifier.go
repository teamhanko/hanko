package registration

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flow_api/services"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
)

// RegisterLoginIdentifier takes the identifier which the user entered and checks if they are valid and available according to the configuration
type RegisterLoginIdentifier struct {
	shared.Action
}

func (a RegisterLoginIdentifier) GetName() flowpilot.ActionName {
	return shared.ActionRegisterLoginIdentifier
}

func (a RegisterLoginIdentifier) GetDescription() string {
	return "Enter an identifier to register."
}

func (a RegisterLoginIdentifier) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Account.AllowSignup {
		c.SuspendAction()
		return
	}

	if (!deps.Cfg.Email.Enabled || (deps.Cfg.Email.Enabled && !deps.Cfg.Email.AcquireOnRegistration)) &&
		(!deps.Cfg.Username.Enabled || (deps.Cfg.Username.Enabled && !deps.Cfg.Username.AcquireOnRegistration)) {
		c.SuspendAction()
		return
	}

	if deps.Cfg.Email.Enabled && deps.Cfg.Email.AcquireOnRegistration {
		input := flowpilot.EmailInput("email").
			MaxLength(deps.Cfg.Email.MaxLength).
			Required(!deps.Cfg.Email.Optional).
			TrimSpace(true).
			LowerCase(true)

		c.AddInputs(input)
	}

	if deps.Cfg.Username.Enabled && deps.Cfg.Username.AcquireOnRegistration {
		input := flowpilot.StringInput("username").
			MinLength(deps.Cfg.Username.MinLength).
			MaxLength(deps.Cfg.Username.MaxLength).
			Required(!deps.Cfg.Username.Optional).
			TrimSpace(true).
			LowerCase(true)

		c.AddInputs(input)
	}
}

func (a RegisterLoginIdentifier) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	email := c.Input().Get("email").String()
	username := c.Input().Get("username").String()

	if deps.Cfg.Email.Optional && len(email) == 0 &&
		deps.Cfg.Username.Optional && len(username) == 0 {
		err := errors.New("either email or username must be provided")
		c.Input().SetError("username", flowpilot.ErrorValueInvalid.Wrap(err))
		c.Input().SetError("email", flowpilot.ErrorValueInvalid.Wrap(err))
		return c.Error(flowpilot.ErrorFormDataInvalid.Wrap(err))
	}

	if username != "" && (deps.Cfg.Username.Enabled && deps.Cfg.Username.AcquireOnRegistration) {
		if !services.ValidateUsername(username) {
			c.Input().SetError("username", shared.ErrorInvalidUsername)
			return c.Error(flowpilot.ErrorFormDataInvalid)
		}

		// Check that username is not already taken (tenant-scoped with fallback to global users)
		// this check is non-exhaustive as the username is not blocked here and might be created after the check here and the user creation
		usernameModel, isGlobalFallback, err := deps.Persister.GetUsernamePersisterWithConnection(deps.Tx).GetByNameWithTenantFallback(username, deps.TenantID)
		if err != nil {
			return err
		}
		if usernameModel != nil {
			// If it's a global user and we have a tenant, they can potentially be adopted during login
			// For registration, we still treat this as "already exists" to redirect to login flow
			if isGlobalFallback && deps.TenantID != nil {
				// Global user exists - they should use login flow to be adopted
				c.Input().SetError("username", shared.ErrorUsernameAlreadyExists)
				return c.Error(flowpilot.ErrorFormDataInvalid)
			}
			c.Input().SetError("username", shared.ErrorUsernameAlreadyExists)
			return c.Error(flowpilot.ErrorFormDataInvalid)
		}

		err = c.CopyInputValuesToStash("username")
		if err != nil {
			return fmt.Errorf("failed to copy username input to the stash: %w", err)
		}
	}

	if email != "" {
		if deps.Cfg.Saml.Enabled {
			domain := strings.Split(email, "@")[1]
			if provider, err := deps.SamlService.GetProviderByDomain(domain); err == nil && provider != nil {
				var authUrl string
				authUrl, err = deps.SamlService.GetAuthUrl(provider, deps.Cfg.Saml.DefaultRedirectUrl, true)

				if err != nil {
					return fmt.Errorf("failed to get auth url: %w", err)
				}

				_ = c.Payload().Set("redirect_url", authUrl)

				return c.Continue(shared.StateThirdParty)
			}
		}

		if deps.Cfg.Email.Enabled && deps.Cfg.Email.AcquireOnRegistration {
			// Check that email is not already taken (tenant-scoped with fallback to global users)
			// this check is non-exhaustive as the email is not blocked here and might be created after the check here and the user creation
			emailModel, isGlobalFallback, err := deps.Persister.GetEmailPersisterWithConnection(deps.Tx).FindByAddressWithTenantFallback(email, deps.TenantID)
			if err != nil {
				return err
			}
			// Do not return an error when only identifier is email and email verification is on (account enumeration protection) and privacy setting is off
			// Note: If a global user exists, they should use the login flow to be adopted into the tenant
			if emailModel != nil {
				// If it's a global user and we have a tenant, they can potentially be adopted during login
				_ = isGlobalFallback // Global users should use login flow to be adopted
				// E-mail address already exists
				if !deps.Cfg.Email.RequireVerification || deps.Cfg.Privacy.ShowAccountExistenceHints {
					c.Input().SetError("email", shared.ErrorEmailAlreadyExists)
					return c.Error(flowpilot.ErrorFormDataInvalid)
				} else {
					err = c.CopyInputValuesToStash("email")
					if err != nil {
						return fmt.Errorf("failed to copy email to stash: %w", err)
					}

					err = c.Stash().Set(shared.StashPathPasscodeTemplate, shared.PasscodeTemplateEmailRegistrationAttempted)
					if err != nil {
						return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
					}

					return c.Continue(shared.StatePasscodeConfirmation)
				}
			}

			err = c.CopyInputValuesToStash("email")
			if err != nil {
				return fmt.Errorf("failed to copy email input to the stash: %w", err)
			}

			if deps.Cfg.Email.RequireVerification {
				if err = c.Stash().Set(shared.StashPathPasscodeTemplate, shared.PasscodeTemplateEmailVerification); err != nil {
					return fmt.Errorf("failed to set passcode_template to stash: %w", err)
				}
			}
		}
	}

	userID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to generate a new user id: %w", err)
	}

	err = c.Stash().Set(shared.StashPathUserID, userID.String())
	if err != nil {
		return fmt.Errorf("failed to stash user_id: %w", err)
	}

	states, err := a.generateRegistrationStates(c)
	if err != nil {
		return err
	}

	return c.Continue(append(states, shared.StateSuccess)...)
}

func (a RegisterLoginIdentifier) generateRegistrationStates(c flowpilot.ExecutionContext) ([]flowpilot.StateName, error) {
	deps := a.GetDeps(c)

	result := make([]flowpilot.StateName, 0)

	if deps.Cfg.Email.Enabled && deps.Cfg.Email.AcquireOnRegistration {
		emailExists := len(c.Input().Get("email").String()) > 0
		if emailExists && deps.Cfg.Email.RequireVerification {
			result = append(result, shared.StatePasscodeConfirmation)
		}
	}

	webauthnAvailable := c.Stash().Get(shared.StashPathWebauthnAvailable).Bool()
	passkeyEnabled := webauthnAvailable && deps.Cfg.Passkey.Enabled
	passwordEnabled := deps.Cfg.Password.Enabled
	passwordAndPasskeyEnabled := passkeyEnabled && passwordEnabled

	alwaysAcquirePasskey := deps.Cfg.Passkey.AcquireOnRegistration == "always"
	conditionalAcquirePasskey := deps.Cfg.Passkey.AcquireOnRegistration == "conditional"
	alwaysAcquirePassword := deps.Cfg.Password.AcquireOnRegistration == "always"
	conditionalAcquirePassword := deps.Cfg.Password.AcquireOnRegistration == "conditional"
	neverAcquirePasskey := deps.Cfg.Passkey.AcquireOnRegistration == "never"
	neverAcquirePassword := deps.Cfg.Password.AcquireOnRegistration == "never"

	if passwordAndPasskeyEnabled {
		if alwaysAcquirePasskey && alwaysAcquirePassword {
			if !deps.Cfg.Password.Optional && deps.Cfg.Passkey.Optional {
				result = append(result, shared.StatePasswordCreation, shared.StateOnboardingCreatePasskey)
			} else {
				result = append(result, shared.StateOnboardingCreatePasskey, shared.StatePasswordCreation)
			}
		} else if alwaysAcquirePasskey && conditionalAcquirePassword {
			result = append(result, shared.StateOnboardingCreatePasskey)
		} else if conditionalAcquirePasskey && alwaysAcquirePassword {
			result = append(result, shared.StatePasswordCreation)
		} else if conditionalAcquirePasskey && conditionalAcquirePassword {
			result = append(result, shared.StateCredentialOnboardingChooser)
		} else if conditionalAcquirePasskey && neverAcquirePassword {
			result = append(result, shared.StateOnboardingCreatePasskey)
		} else if neverAcquirePasskey && (alwaysAcquirePassword || conditionalAcquirePassword) {
			result = append(result, shared.StatePasswordCreation)
		} else if (alwaysAcquirePasskey || conditionalAcquirePasskey) && neverAcquirePassword {
			result = append(result, shared.StateOnboardingCreatePasskey)
		}
	} else if passkeyEnabled && (alwaysAcquirePasskey || conditionalAcquirePasskey) {
		result = append(result, shared.StateOnboardingCreatePasskey)
	} else if passwordEnabled && (alwaysAcquirePassword || conditionalAcquirePassword) {
		result = append(result, shared.StatePasswordCreation)
	}

	if len(result) == 0 {
		err := c.ExecuteHook(shared.ScheduleMFACreationStates{})
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
