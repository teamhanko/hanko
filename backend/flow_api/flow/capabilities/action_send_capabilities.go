package capabilities

import (
	"github.com/teamhanko/hanko/backend/flow_api/constants"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type RegisterClientCapabilities struct {
	shared.Action
}

func (a RegisterClientCapabilities) GetName() flowpilot.ActionName {
	return constants.ActionRegisterClientCapabilities
}

func (a RegisterClientCapabilities) GetDescription() string {
	return "Send the computers capabilities."
}

func (a RegisterClientCapabilities) Initialize(c flowpilot.InitializationContext) {
	//deps := a.GetDeps(c)
	//
	//if !deps.Cfg.Passkey.Enabled {
	//	c.SuspendAction()
	//	return
	//}

	c.AddInputs(flowpilot.BooleanInput("webauthn_conditional_mediation_available").
		Hidden(true))
	c.AddInputs(flowpilot.BooleanInput("webauthn_available").Required(true).Hidden(true))
}

func (a RegisterClientCapabilities) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	//if !deps.Cfg.Passkey.Enabled {
	//	c.ContinueFlow()
	//}

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	webauthnAvailable := c.Input().Get("webauthn_available").String() == "true"

	// Only passkeys are allowed, but webauthn is not available on the browser
	if !webauthnAvailable && !deps.Cfg.Password.Enabled && !deps.Cfg.Passcode.Enabled {
		return c.ContinueFlowWithError(constants.StateError, shared.ErrorDeviceNotCapable)
	}

	// Only security keys are allowed as a second factor, but webauthn is not available on the browser
	if !webauthnAvailable &&
		deps.Cfg.SecondFactor.Enabled && !deps.Cfg.SecondFactor.Optional &&
		len(deps.Cfg.SecondFactor.Methods) == 1 &&
		deps.Cfg.SecondFactor.Methods[0] == "security_key" {
		return c.ContinueFlowWithError(constants.StateError, shared.ErrorDeviceNotCapable)
	}

	err := c.Stash().Set("webauthn_available", webauthnAvailable)
	if err != nil {
		return err
	}

	conditionalMediationAvailable := c.Input().Get("webauthn_conditional_mediation_available").Bool()
	err = c.Stash().Set("webauthn_conditional_mediation_available", conditionalMediationAvailable)
	if err != nil {
		return err
	}

	return c.EndSubFlow()
}

func (a RegisterClientCapabilities) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
