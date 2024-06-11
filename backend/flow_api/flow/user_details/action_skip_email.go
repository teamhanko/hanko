package user_details

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type SkipEmail struct {
	shared.Action
}

func (a SkipEmail) GetName() flowpilot.ActionName {
	return shared.ActionSkip
}

func (a SkipEmail) GetDescription() string {
	return "Skip"
}

func (a SkipEmail) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Email.Optional {
		c.SuspendAction()
	}
}

func (a SkipEmail) Execute(c flowpilot.ExecutionContext) error {
	if err := c.Stash().Set("suspend_back_action", false); err != nil {
		return fmt.Errorf("failed to set suspend_back_action to the stash: %w", err)
	}

	return c.EndSubFlow()

}
