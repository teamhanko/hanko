package passkey_onboarding

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type Skip struct {
	shared.Action
}

func (a Skip) GetName() flowpilot.ActionName {
	return ActionSkip
}

func (a Skip) GetDescription() string {
	return "Skip the passkey onboarding"
}

func (a Skip) Initialize(c flowpilot.InitializationContext) {
	if !c.Stash().Get("allow_skip_onboarding").Bool() {
		c.SuspendAction()
	}
}

func (a Skip) Execute(c flowpilot.ExecutionContext) error {
	return c.EndSubFlow()
}
