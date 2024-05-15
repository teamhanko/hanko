package registration_method_chooser

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

	if !deps.Cfg.Email.RequireVerification {
		c.SuspendAction()
	}
}

func (a Skip) Execute(c flowpilot.ExecutionContext) error {
	return c.EndSubFlow()
}

func (a Skip) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
