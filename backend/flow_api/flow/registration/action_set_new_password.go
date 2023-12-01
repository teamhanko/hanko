package registration

import (
	"errors"
	"fmt"
	passkeyOnboarding "github.com/teamhanko/hanko/backend/flow_api/flow/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"golang.org/x/crypto/bcrypt"
	"unicode/utf8"
)

type SetNewPassword struct {
	shared.Action
}

func (a SetNewPassword) GetName() flowpilot.ActionName {
	return ActionSetNewPassword
}

func (a SetNewPassword) GetDescription() string {
	return "Submit a new password."
}

func (a SetNewPassword) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)
	c.AddInputs(flowpilot.PasswordInput("new_password").Required(true).MinLength(deps.Cfg.Password.MinPasswordLength))
}

func (a SetNewPassword) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	newPassword := c.Input().Get("new_password").String()
	newPasswordBytes := []byte(newPassword)

	if utf8.RuneCountInString(newPassword) < deps.Cfg.Password.MinPasswordLength {
		c.Input().SetError("new_password", flowpilot.ErrorValueInvalid)
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(newPasswordBytes, 12)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			c.Input().SetError("new_password", flowpilot.ErrorValueTooLong)
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		}
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = c.Stash().Set("new_password", string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to set new_password to stash: %w", err)
	}

	// Decide which is the next state according to the config and user input
	if deps.Cfg.Passkey.Onboarding.Enabled && c.Stash().Get("webauthn_available").Bool() {
		return c.StartSubFlow(passkeyOnboarding.StateIntroduction, StateSuccess)
	}

	return c.ContinueFlow(StateSuccess)
}
