package registration

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"unicode/utf8"
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

	if (!deps.Cfg.Email.Enabled || (deps.Cfg.Email.Enabled && !deps.Cfg.Email.AcquireOnRegistration)) &&
		(!deps.Cfg.Username.Enabled || (deps.Cfg.Username.Enabled && !deps.Cfg.Username.AcquireOnRegistration)) {
		c.SuspendAction()
		return
	}

	if deps.Cfg.Email.Enabled && deps.Cfg.Email.AcquireOnRegistration {
		input := flowpilot.EmailInput("email").
			MaxLength(255).
			Persist(true).
			Preserve(true).
			Required(!deps.Cfg.Email.Optional)

		c.AddInputs(input)
	}

	if deps.Cfg.Username.Enabled && deps.Cfg.Username.AcquireOnRegistration {
		input := flowpilot.StringInput("username").
			MinLength(deps.Cfg.Username.MinLength).
			MaxLength(deps.Cfg.Username.MaxLength).
			Persist(true).
			Preserve(true).
			Required(!deps.Cfg.Username.Optional)

		c.AddInputs(input)
	}
}

func (a RegisterLoginIdentifier) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	email := c.Input().Get("email").String()
	username := c.Input().Get("username").String()

	// check that username only contains allowed characters
	if !utf8.ValidString(username) {
		c.Input().SetError("username", flowpilot.ErrorValueInvalid.Wrap(errors.New("username contains invalid characters")))
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	if username != "" {
		// Check that username is not already taken
		// this check is non-exhaustive as the username is not blocked here and might be created after the check here and the user creation
		userModel, err := deps.Persister.GetUserPersister().GetByUsername(username)
		if err != nil {
			return err
		}
		if userModel != nil {
			c.Input().SetError("username", shared.ErrorUsernameAlreadyExists)
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		}
	}

	if email != "" {
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
				return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
			} else {
				err = c.CopyInputValuesToStash("email")
				if err != nil {
					return fmt.Errorf("failed to copy email to stash: %w", err)
				}

				err = c.Stash().Set("passcode_template", "email_registration_attempted")
				if err != nil {
					return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
				}

				return c.StartSubFlow(shared.StatePasscodeConfirmation)
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

	err = c.Stash().Set("user_id", userID.String())
	if err != nil {
		return fmt.Errorf("failed to stash user_id: %w", err)
	}

	doEmailVerification := email != "" && deps.Cfg.Email.RequireVerification

	if doEmailVerification {
		if err := c.Stash().Set("passcode_template", "email_verification"); err != nil {
			return fmt.Errorf("failed to set passcode_template to stash: %w", err)
		}
	}

	states, err := a.generateRegistrationStates(deps.Cfg, generateRegistrationStatesParams{
		WebauthnAvailable: c.Stash().Get("webauthn_available").Bool(),
		EmailVerification: doEmailVerification,
	})

	if err != nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFlowDiscontinuity.Wrap(err))
	}

	return c.StartSubFlow(states[0], states[1:]...)
}

func (a RegisterLoginIdentifier) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}

type generateRegistrationStatesParams struct {
	WebauthnAvailable bool
	EmailVerification bool
}

func (a RegisterLoginIdentifier) generateRegistrationStates(cfg config.Config, p generateRegistrationStatesParams) ([]flowpilot.StateName, error) {
	stateNames := make([]flowpilot.StateName, 0)

	if p.EmailVerification {
		stateNames = append(stateNames, shared.StatePasscodeConfirmation)
	}

	passkeyEnabled := p.WebauthnAvailable && cfg.Passkey.Enabled
	passwordEnabled := cfg.Password.Enabled
	bothEnabled := passkeyEnabled && passwordEnabled

	passkeyOptional := cfg.Passkey.Optional
	passwordOptional := cfg.Password.Optional
	bothOptional := passwordOptional && passkeyOptional

	acquirePassword := cfg.Password.AcquireOnRegistration
	acquirePasskey := cfg.Passkey.AcquireOnRegistration
	acquireBoth := acquirePassword && acquirePasskey

	if bothEnabled {
		if acquireBoth {
			if bothOptional {
				// Wenn !p.EmailVerification, dann darf man den Chooser nicht skippen
				stateNames = append(stateNames, shared.StateCredentialOnboardingChooser)
			} else if passwordOptional {
				// Man darf Password skippen
				stateNames = append(stateNames, shared.StateOnboardingCreatePasskey, shared.StatePasswordCreation)
			} else if passkeyOptional {
				// Man darf Passkey skippen
				stateNames = append(stateNames, shared.StatePasswordCreation, shared.StateOnboardingCreatePasskey)
			} else {
				// Man darf keine der beiden Methoden skippen
				stateNames = append(stateNames, shared.StateOnboardingCreatePasskey, shared.StatePasswordCreation)
			}
		} else if acquirePassword {
			// Wenn !p.EmailVerification, dann darf man Password Creation nicht skippen
			stateNames = append(stateNames, shared.StatePasswordCreation)
		} else if acquirePasskey {
			stateNames = append(stateNames, shared.StateOnboardingCreatePasskey)
		} else {
			return nil, errors.New("no credential acquired")
		}
	} else if passkeyEnabled {
		if acquirePasskey {
			stateNames = append(stateNames, shared.StateOnboardingCreatePasskey)
		} else {
			return nil, errors.New("no credential acquired")
		}
	} else if passwordEnabled {
		if acquirePassword {
			stateNames = append(stateNames, shared.StatePasswordCreation)
		} else {
			return nil, errors.New("no credential acquired")
		}
	} else if p.EmailVerification {
		return append(stateNames, shared.StateSuccess), nil
	} else {
		return nil, errors.New("no credential acquired")
	}

	return append(stateNames, shared.StateSuccess), nil
}
