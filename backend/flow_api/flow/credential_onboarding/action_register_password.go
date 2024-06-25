package credential_onboarding

import (
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"golang.org/x/crypto/bcrypt"
	"unicode/utf8"
)

type RegisterPassword struct {
	shared.Action
}

func (a RegisterPassword) GetName() flowpilot.ActionName {
	return shared.ActionRegisterPassword
}

func (a RegisterPassword) GetDescription() string {
	return "Submit a new password."
}

func (a RegisterPassword) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)
	c.AddInputs(flowpilot.PasswordInput("new_password").Required(!deps.Cfg.Password.Optional).MinLength(deps.Cfg.Password.MinLength))
}

func (a RegisterPassword) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	newPassword := c.Input().Get("new_password").String()
	newPasswordBytes := []byte(newPassword)

	if utf8.RuneCountInString(newPassword) < deps.Cfg.Password.MinLength {
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

	err = c.Stash().Set(shared.StashPathNewPassword, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to set new_password to stash: %w", err)
	}

	err = c.Stash().Set(shared.StashPathUserHasPassword, true)
	if err != nil {
		return fmt.Errorf("failed to set user_has_password to the stash: %w", err)
	}

	err = c.DeleteStateHistory(true)
	if err != nil {
		return fmt.Errorf("failed to delete state history: %w", err)
	}

	return c.EndSubFlow()
}
