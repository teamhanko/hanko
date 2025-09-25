package credential_usage

import (
	"fmt"

	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
)

type RememberMe struct {
	shared.Action
}

func (a RememberMe) GetName() flowpilot.ActionName {
	return shared.ActionRememberMe
}

func (a RememberMe) GetDescription() string {
	return "Enables the user to stay signed in."
}

func (a RememberMe) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	c.AddInputs(flowpilot.BooleanInput("remember_me").Required(true))

	if deps.Cfg.Session.Cookie.Retention != "prompt" {
		c.SuspendAction()
	}
}

func (a RememberMe) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	rememberMeSelected := c.Input().Get("remember_me").Bool()

	if err := c.Stash().Set(shared.StashPathRememberMeSelected, rememberMeSelected); err != nil {
		return fmt.Errorf("failed to set remember_me_selected to stash: %w", err)
	}

	return c.Continue(c.GetCurrentState())
}
