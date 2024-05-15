package register_password

import (
	"github.com/teamhanko/hanko/backend/flow_api/constants"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type Skip struct {
	shared.Action
}

func (a Skip) GetName() flowpilot.ActionName {
	return constants.ActionSkip
}

func (a Skip) GetDescription() string {
	return "Skip"
}

func (a Skip) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Password.Optional || !deps.Cfg.Email.RequireVerification {
		c.SuspendAction()
	}
}
func (a Skip) Execute(c flowpilot.ExecutionContext) error {
	switch c.GetFlowName() {
	case "registration":
		return c.EndSubFlow()
	case "login":
		return c.ContinueFlow(constants.StateOnboardingCreatePasskeyConditional)
	}
	return nil
}

func (a Skip) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
