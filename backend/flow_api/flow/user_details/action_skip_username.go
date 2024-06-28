package user_details

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type SkipUsername struct {
	shared.Action
}

func (a SkipUsername) GetName() flowpilot.ActionName {
	return shared.ActionSkip
}

func (a SkipUsername) GetDescription() string {
	return "Skip"
}

func (a SkipUsername) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Username.Optional {
		c.SuspendAction()
	}
}
func (a SkipUsername) Execute(c flowpilot.ExecutionContext) error {
	if err := c.DeleteStateHistory(true); err != nil {
		return fmt.Errorf("failed to delete the state history: %w", err)
	}

	return c.Continue()
}
