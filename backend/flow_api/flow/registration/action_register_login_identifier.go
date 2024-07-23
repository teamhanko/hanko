package registration

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"strings"
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

	if username != "" {
		if !services.ValidateUsername(username) {
			c.Input().SetError("username", shared.ErrorInvalidUsername)
			return c.Error(flowpilot.ErrorFormDataInvalid)
		}

		// Check that username is not already taken
		// this check is non-exhaustive as the username is not blocked here and might be created after the check here and the user creation
		userModel, err := deps.Persister.GetUserPersister().GetByUsername(username)
		if err != nil {
			return err
		}
		if userModel != nil {
			c.Input().SetError("username", shared.ErrorUsernameAlreadyExists)
			return c.Error(flowpilot.ErrorFormDataInvalid)
		}
	}

	if email != "" {
		if deps.Cfg.Saml.Enabled {
			domain := strings.Split(email, "@")[1]
			if provider, err := deps.SamlService.GetProviderByDomain(domain); err == nil && provider != nil {
				authUrl, err := deps.SamlService.GetAuthUrl(provider, deps.Cfg.Saml.DefaultRedirectUrl, true)

				if err != nil {
					return fmt.Errorf("failed to get auth url: %w", err)
				}

				_ = c.Payload().Set("redirect_url", authUrl)

				return c.Continue(shared.StateThirdParty)
			}
		}

		// Check that email is not already taken
		// this check is non-exhaustive as the email is not blocked here and might be created after the check here and the user creation
		emailModel, err := deps.Persister.GetEmailPersister().FindByAddress(email)
		if err != nil {
			return err
		}
		// Do not return an error when only identifier is email and email verification is on (account enumeration protection)
		if emailModel != nil {
			// E-mail address already exists
			if !deps.Cfg.Email.RequireVerification {
				c.Input().SetError("email", shared.ErrorEmailAlreadyExists)
				return c.Error(flowpilot.ErrorFormDataInvalid)
			} else {
				err = c.CopyInputValuesToStash("email")
				if err != nil {
					return fmt.Errorf("failed to copy email to stash: %w", err)
				}

				err = c.Stash().Set(shared.StashPathPasscodeTemplate, "email_registration_attempted")
				if err != nil {
					return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
				}

				return c.Continue(shared.StatePasscodeConfirmation)
			}
		}
	}

	err := c.CopyInputValuesToStash("email", "username")
	if err != nil {
		return fmt.Errorf("failed to copy input values to the stash: %w", err)
	}

	userID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to generate a new user id: %w", err)
	}

	err = c.Stash().Set(shared.StashPathUserID, userID.String())
	if err != nil {
		return fmt.Errorf("failed to stash user_id: %w", err)
	}

	if email != "" && deps.Cfg.Email.RequireVerification {
		if err = c.Stash().Set(shared.StashPathPasscodeTemplate, "email_verification"); err != nil {
			return fmt.Errorf("failed to set passcode_template to stash: %w", err)
		}
	}

	return c.Continue(append(a.generateRegistrationStates(c), shared.StateSuccess)...)
}

func (a RegisterLoginIdentifier) generateRegistrationStates(c flowpilot.ExecutionContext) []flowpilot.StateName {
	deps := a.GetDeps(c)

	stateNames := make([]flowpilot.StateName, 0)

	emailExists := len(c.Input().Get("email").String()) > 0
	if emailExists && deps.Cfg.Email.RequireVerification {
		stateNames = append(stateNames, shared.StatePasscodeConfirmation)
	}

	webauthnAvailable := c.Stash().Get(shared.StashPathWebauthnAvailable).Bool()
	passkeyEnabled := webauthnAvailable && deps.Cfg.Passkey.Enabled
	passwordEnabled := deps.Cfg.Password.Enabled
	bothEnabled := passkeyEnabled && passwordEnabled

	alwaysPasskey := deps.Cfg.Passkey.AcquireOnRegistration == "always"
	conditionalPasskey := deps.Cfg.Passkey.AcquireOnRegistration == "conditional"
	alwaysPassword := deps.Cfg.Password.AcquireOnRegistration == "always"
	conditionalPassword := deps.Cfg.Password.AcquireOnRegistration == "conditional"

	if bothEnabled {
		if conditionalPasskey && conditionalPassword {
			stateNames = append(stateNames, shared.StateCredentialOnboardingChooser)
		} else if alwaysPasskey && !alwaysPassword {
			stateNames = append(stateNames, shared.StateOnboardingCreatePasskey)
		} else if !alwaysPasskey && alwaysPassword {
			stateNames = append(stateNames, shared.StatePasswordCreation)
		} else if alwaysPassword && alwaysPasskey {
			if !deps.Cfg.Password.Optional && deps.Cfg.Passkey.Optional {
				stateNames = append(stateNames, shared.StatePasswordCreation, shared.StateOnboardingCreatePasskey)
			} else {
				stateNames = append(stateNames, shared.StateOnboardingCreatePasskey, shared.StatePasswordCreation)
			}
		}
	} else if passkeyEnabled && (alwaysPasskey || conditionalPasskey) {
		stateNames = append(stateNames, shared.StateOnboardingCreatePasskey)
	} else if passwordEnabled && (alwaysPassword || conditionalPassword) {
		stateNames = append(stateNames, shared.StatePasswordCreation)
	}

	return stateNames
}
