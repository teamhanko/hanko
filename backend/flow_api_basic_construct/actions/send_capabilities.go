package actions

import (
	"encoding/json"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
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
	return common.ActionSendCapabilities
}

func (m SendCapabilities) GetDescription() string {
	return "Send the computers capabilities."
}

func (m SendCapabilities) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("capabilities").Required(true).Hidden(true))
}

func (m SendCapabilities) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	var capabilities capabilities
	err := json.Unmarshal([]byte(c.Input().Get("capabilities").String()), &capabilities)
	if err != nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
	}

	if capabilities.Webauthn.Available == false &&
		m.cfg.Password.Enabled == false &&
		m.cfg.Passcode.Enabled == false {
		// Only passkeys are allowed, but webauthn is not available on the browser
		return c.ContinueFlowWithError(common.StateError, common.ErrorDeviceNotCapable)
	}
	if capabilities.Webauthn.Available == false &&
		m.cfg.SecondFactor.Enabled == "required" &&
		len(m.cfg.SecondFactor.Methods) == 1 &&
		m.cfg.SecondFactor.Methods[0] == "security_key" {
		// Only security keys are allowed as a second factor, but webauthn is not available on the browser
		return c.ContinueFlowWithError(common.StateError, common.ErrorDeviceNotCapable)
	}

	err = c.Stash().Set("capabilities", capabilities)
	if err != nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
	}

	if c.GetCurrentState() == common.StateRegistrationPreflight {
		return c.ContinueFlow(common.StateRegistrationInit)
	} else {
		return c.ContinueFlow(common.StateLoginInit)
	}
}

type capabilities struct {
	Webauthn webauthn `json:"webauthn"`
}

type webauthn struct {
	Available bool `json:"available"`
}