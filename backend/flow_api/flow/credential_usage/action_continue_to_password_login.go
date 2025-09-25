package credential_usage

import (
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
)

type ContinueToPasswordLogin struct {
	shared.Action
}

func (a ContinueToPasswordLogin) GetName() flowpilot.ActionName {
	return shared.ActionContinueToPasswordLogin
}

func (a ContinueToPasswordLogin) GetDescription() string {
	return "Continue to the password login."
}

func (a ContinueToPasswordLogin) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)
	if deps.Cfg.Privacy.OnlyShowActualLoginMethods && (!c.Stash().Get(shared.StashPathUserHasPassword).Bool() || !deps.Cfg.Password.Enabled) {
		c.SuspendAction()
	}
}

func (a ContinueToPasswordLogin) Execute(c flowpilot.ExecutionContext) error {
	return c.Continue(shared.StateLoginPassword)
}
