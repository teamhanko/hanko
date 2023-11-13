package actions

import (
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"golang.org/x/crypto/bcrypt"
	"unicode/utf8"
)

func NewSubmitNewPassword(cfg config.Config) SubmitNewPassword {
	return SubmitNewPassword{
		cfg,
	}
}

type SubmitNewPassword struct {
	cfg config.Config
}

func (m SubmitNewPassword) GetName() flowpilot.ActionName {
	return common.ActionSubmitNewPassword
}

func (m SubmitNewPassword) GetDescription() string {
	return "Submit a new password."
}

func (m SubmitNewPassword) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.PasswordInput("new_password").Required(true).MinLength(m.cfg.Password.MinPasswordLength))
}

func (m SubmitNewPassword) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	newPassword := c.Input().Get("new_password").String()
	newPasswordBytes := []byte(newPassword)
	if utf8.RuneCountInString(newPassword) < m.cfg.Password.MinPasswordLength {
		c.Input().SetError("new_password", flowpilot.ErrorValueInvalid)
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	if len(newPasswordBytes) > 72 {
		c.Input().SetError("new_password", flowpilot.ErrorValueTooLong)
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(c.Input().Get("new_password").String()), 12)
	if err != nil {
		return err
	}
	err = c.Stash().Set("new_password", string(hashedPassword))

	// Decide which is the next state according to the config and user input
	if m.cfg.Passkey.Onboarding.Enabled && c.Stash().Get("webauthn_available").Bool() {
		return c.StartSubFlow(common.StateOnboardingCreatePasskey, common.StateSuccess)
	}
	// TODO: 2FA routing

	return c.ContinueFlow(common.StateSuccess)
}
