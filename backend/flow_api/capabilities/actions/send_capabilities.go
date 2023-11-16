package actions

import (
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewSendCapabilities(cfg config.Config) SendCapabilities {
	return SendCapabilities{
		cfg,
	}
}

type SendCapabilities struct {
	cfg config.Config
}

func (m SendCapabilities) GetName() flowpilot.ActionName {
	return shared.ActionSendCapabilities
}

func (m SendCapabilities) GetDescription() string {
	return "Send the computers capabilities."
}

func (m SendCapabilities) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("webauthn_available").Required(true).Hidden(true))
}

func (m SendCapabilities) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	webauthnAvailable := c.Input().Get("webauthn_available").String() == "true"

	// Only passkeys are allowed, but webauthn is not available on the browser
	if !webauthnAvailable && !m.cfg.Password.Enabled && !m.cfg.Passcode.Enabled {
		return c.ContinueFlowWithError(shared.StateError, shared.ErrorDeviceNotCapable)
	}

	// Only security keys are allowed as a second factor, but webauthn is not available on the browser
	if !webauthnAvailable &&
		m.cfg.SecondFactor.Enabled && !m.cfg.SecondFactor.Optional &&
		len(m.cfg.SecondFactor.Methods) == 1 &&
		m.cfg.SecondFactor.Methods[0] == "security_key" {
		return c.ContinueFlowWithError(shared.StateError, shared.ErrorDeviceNotCapable)
	}

	err := c.Stash().Set("webauthn_available", webauthnAvailable)
	if err != nil {
		return err
	}

	return c.EndSubFlow()
}
