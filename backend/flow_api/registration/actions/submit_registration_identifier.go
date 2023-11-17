package actions

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	passcodeStates "github.com/teamhanko/hanko/backend/flow_api/passcode/states"
	passkeyOnboardingStates "github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding/states"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flow_api/shared/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"strings"
)

// SubmitRegistrationIdentifier takes the identifier which the user entered and checks if they are valid and available according to the configuration
type SubmitRegistrationIdentifier struct {
	cfg             config.Config
	persister       persistence.Persister
	httpContext     echo.Context
	passcodeService services.Passcode
}

func (m SubmitRegistrationIdentifier) GetName() flowpilot.ActionName {
	return shared.ActionSubmitRegistrationIdentifier
}

func (m SubmitRegistrationIdentifier) GetDescription() string {
	return "Enter an identifier to register."
}

func (m SubmitRegistrationIdentifier) Initialize(c flowpilot.InitializationContext) {
	if m.cfg.Identifier.Email.Enabled {
		input := flowpilot.EmailInput("email").
			MaxLength(255).
			Persist(true).
			Preserve(true).
			Required(!m.cfg.Identifier.Email.Optional)

		c.AddInputs(input)
	}

	if m.cfg.Identifier.Username.Enabled {
		input := flowpilot.StringInput("username").
			MinLength(m.cfg.Identifier.Username.MinLength).
			MaxLength(m.cfg.Identifier.Username.MaxLength).
			Persist(true).
			Preserve(true).
			Required(!m.cfg.Identifier.Username.Optional)

		c.AddInputs(input)
	}
}

func (m SubmitRegistrationIdentifier) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	email := c.Input().Get("email").String()
	username := c.Input().Get("username").String()

	for _, char := range username {
		// check that username only contains allowed characters
		if !strings.Contains(m.cfg.Identifier.Username.AllowedCharacters, string(char)) {
			c.Input().SetError("username", flowpilot.ErrorValueInvalid.Wrap(errors.New("username contains invalid characters")))
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		}
	}

	if email != "" {
		// Check that email is not already taken
		// this check is non-exhaustive as the email is not blocked here and might be created after the check here and the user creation
		e, err := m.persister.GetEmailPersister().FindByAddress(email)
		if err != nil {
			return err
		}
		// Do not return an error when only identifier is email and email verification is on (account enumeration protection)
		if e != nil && !(!m.cfg.Identifier.Username.Enabled && m.cfg.Emails.RequireVerification) {
			c.Input().SetError("email", shared.ErrorEmailAlreadyExists)
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		}
	}

	if username != "" {
		// Check that username is not already taken
		// this check is non-exhaustive as the username is not blocked here and might be created after the check here and the user creation
		u, err := m.persister.GetUserPersister().GetByUsername(username)
		if err != nil {
			return err
		}
		if u != nil {
			c.Input().SetError("username", shared.ErrorUsernameAlreadyExists)
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		}
	}

	err := c.CopyInputValuesToStash("email", "username")
	if err != nil {
		return err
	}

	// Decide which is the next state according to the config and user input
	if email != "" && m.cfg.Emails.RequireVerification {
		if err := c.Stash().Set("passcode_template", "email_verification"); err != nil {
			return fmt.Errorf("failed to set passcode_template to stash: %w", err)
		}

		if m.cfg.Password.Enabled {
			return c.StartSubFlow(passcodeStates.StatePasscodeConfirmation, shared.StatePasswordCreation)
		} else if !m.cfg.Passcode.Enabled || (m.cfg.Passkey.Onboarding.Enabled && c.Stash().Get("webauthn_available").Bool()) {
			return c.StartSubFlow(passcodeStates.StatePasscodeConfirmation, passkeyOnboardingStates.StateOnboardingCreatePasskey, shared.StateSuccess)
		}
	} else if m.cfg.Password.Enabled {
		return c.ContinueFlow(shared.StatePasswordCreation)
	} else if !m.cfg.Passcode.Enabled {
		return c.StartSubFlow(passkeyOnboardingStates.StateOnboardingCreatePasskey, shared.StateSuccess)
	}

	// TODO: store user and create session token // should this case even exist (only works when email (optional/required) is set by the user) ???

	return c.ContinueFlow(shared.StateSuccess)
}
