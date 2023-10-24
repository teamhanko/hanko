package actions

import (
	"errors"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewSkip(cfg config.Config) Skip {
	return Skip{
		cfg,
	}
}

type Skip struct {
	cfg config.Config
}

func (m Skip) GetName() flowpilot.ActionName {
	return common.ActionSkip
}

func (m Skip) GetDescription() string {
	return "Skip the current state."
}

func (m Skip) Initialize(c flowpilot.InitializationContext) {
	if !m.cfg.Passcode.Enabled && !m.cfg.Password.Enabled {
		// suspend action when only passkeys are allowed
		c.SuspendAction()
	}
}

func (m Skip) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	switch c.GetCurrentState() {
	case common.StateOnboardingCreatePasskey:
		return c.EndSubFlow()
	default:
		// return an error, so we don't implicitly continue to unwanted state
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(errors.New("no destination is defined")))
	}
}
