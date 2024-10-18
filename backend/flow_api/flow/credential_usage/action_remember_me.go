package credential_usage

import (
	"fmt"

	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type RememberMe struct {
	shared.Action
}

func (a RememberMe) GetName() flowpilot.ActionName {
	return shared.ActionRememberMe
}

func (a RememberMe) GetDescription() string {
	return "Show a remember me checkbox."
}

func (a RememberMe) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	c.AddInputs(flowpilot.BooleanInput("remember_me").Required(true))

	if deps.Cfg.Session.EnableRememberMe {
		c.SuspendAction()
	}
}

func (a RememberMe) Execute(c flowpilot.ExecutionContext) error {

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	if err := c.Stash().Set(shared.StashPathRememberMe, c.Input().Get(shared.StashPathRememberMe).Bool()); err != nil {
		return fmt.Errorf("failed to set remember_me to stash: %w", err)
	}

	return c.Continue(shared.StateLoginInit)
}
