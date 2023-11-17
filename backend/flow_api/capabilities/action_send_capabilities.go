package capabilities

import (
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type SendCapabilities struct {
	shared.Action
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
	deps := m.GetDepsForExecution(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	webauthnAvailable := c.Input().Get("webauthn_available").String() == "true"

	// Only passkeys are allowed, but webauthn is not available on the browser
	if !webauthnAvailable && !deps.Cfg.Password.Enabled && !deps.Cfg.Passcode.Enabled {
		return c.ContinueFlowWithError(shared.StateError, shared.ErrorDeviceNotCapable)
	}

	// Only security keys are allowed as a second factor, but webauthn is not available on the browser
	if !webauthnAvailable &&
		deps.Cfg.SecondFactor.Enabled && !deps.Cfg.SecondFactor.Optional &&
		len(deps.Cfg.SecondFactor.Methods) == 1 &&
		deps.Cfg.SecondFactor.Methods[0] == "security_key" {
		return c.ContinueFlowWithError(shared.StateError, shared.ErrorDeviceNotCapable)
	}

	err := c.Stash().Set("webauthn_available", webauthnAvailable)
	if err != nil {
		return err
	}

	return c.EndSubFlow()
}
