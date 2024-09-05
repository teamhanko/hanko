package mfa_creation

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type SkipMFA struct {
	shared.Action
}

func (a SkipMFA) GetName() flowpilot.ActionName {
	return shared.ActionSkip
}

func (a SkipMFA) GetDescription() string {
	return "Skip"
}

func (a SkipMFA) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.MFA.Optional {
		c.SuspendAction()
	}
}
func (a SkipMFA) Execute(c flowpilot.ExecutionContext) error {
	return c.Continue()
}
