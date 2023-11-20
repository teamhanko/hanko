package registration

import (
	"errors"
	"fmt"
	passcode "github.com/teamhanko/hanko/backend/flow_api/passcode"
	passkeyOnboarding "github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"strings"
)

// SubmitRegistrationIdentifier takes the identifier which the user entered and checks if they are valid and available according to the configuration
type SubmitRegistrationIdentifier struct {
	shared.Action
}

func (a SubmitRegistrationIdentifier) GetName() flowpilot.ActionName {
	return ActionSubmitRegistrationIdentifier
}

func (a SubmitRegistrationIdentifier) GetDescription() string {
	return "Enter an identifier to register."
}

func (a SubmitRegistrationIdentifier) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if deps.Cfg.Identifier.Email.Enabled {
		input := flowpilot.EmailInput("email").
			MaxLength(255).
			Persist(true).
			Preserve(true).
			Required(!deps.Cfg.Identifier.Email.Optional)

		c.AddInputs(input)
	}

	if deps.Cfg.Identifier.Username.Enabled {
		input := flowpilot.StringInput("username").
			MinLength(deps.Cfg.Identifier.Username.MinLength).
			MaxLength(deps.Cfg.Identifier.Username.MaxLength).
			Persist(true).
			Preserve(true).
			Required(!deps.Cfg.Identifier.Username.Optional)

		c.AddInputs(input)
	}
}

func (a SubmitRegistrationIdentifier) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	email := c.Input().Get("email").String()
	username := c.Input().Get("username").String()

	for _, char := range username {
		// check that username only contains allowed characters
		if !strings.Contains(deps.Cfg.Identifier.Username.AllowedCharacters, string(char)) {
			c.Input().SetError("username", flowpilot.ErrorValueInvalid.Wrap(errors.New("username contains invalid characters")))
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		}
	}

	if email != "" {
		// Check that email is not already taken
		// this check is non-exhaustive as the email is not blocked here and might be created after the check here and the user creation
		e, err := deps.Persister.GetEmailPersister().FindByAddress(email)
		if err != nil {
			return err
		}
		// Do not return an error when only identifier is email and email verification is on (account enumeration protection)
		if e != nil && !(!deps.Cfg.Identifier.Username.Enabled && deps.Cfg.Emails.RequireVerification) {
			c.Input().SetError("email", shared.ErrorEmailAlreadyExists)
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		}
	}

	if username != "" {
		// Check that username is not already taken
		// this check is non-exhaustive as the username is not blocked here and might be created after the check here and the user creation
		u, err := deps.Persister.GetUserPersister().GetByUsername(username)
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
	if email != "" && deps.Cfg.Emails.RequireVerification {
		if err := c.Stash().Set("passcode_template", "email_verification"); err != nil {
			return fmt.Errorf("failed to set passcode_template to stash: %w", err)
		}

		if deps.Cfg.Password.Enabled {
			return c.StartSubFlow(passcode.StatePasscodeConfirmation, StatePasswordCreation)
		} else if !deps.Cfg.Passcode.Enabled || (deps.Cfg.Passkey.Onboarding.Enabled && c.Stash().Get("webauthn_available").Bool()) {
			return c.StartSubFlow(passcode.StatePasscodeConfirmation, passkeyOnboarding.StateOnboardingCreatePasskey, shared.StateSuccess)
		}
	} else if deps.Cfg.Password.Enabled {
		return c.ContinueFlow(StatePasswordCreation)
	} else if !deps.Cfg.Passcode.Enabled {
		return c.StartSubFlow(passkeyOnboarding.StateOnboardingCreatePasskey, shared.StateSuccess)
	}

	// TODO: store user and create session token // should this case even exist (only works when email (optional/required) is set by the user) ???

	return c.ContinueFlow(shared.StateSuccess)
}
