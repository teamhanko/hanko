package login

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPasswordLogin struct {
	shared.Action
}

func (a ContinueToPasswordLogin) GetName() flowpilot.ActionName {
	return ActionContinueToPasswordLogin
}

func (a ContinueToPasswordLogin) GetDescription() string {
	return "Continue to the password login."
}

func (a ContinueToPasswordLogin) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Password.Enabled {
		c.SuspendAction()
	}
}

func (a ContinueToPasswordLogin) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueFlow(StateLoginPassword)
}
