package shared

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type Skip struct {
	Action
}

func (a Skip) GetName() flowpilot.ActionName {
	return ActionSkip
}

func (a Skip) GetDescription() string {
	return "Skip"
}

func (a Skip) Initialize(c flowpilot.InitializationContext) {}

func (a Skip) Execute(c flowpilot.ExecutionContext) error {
	return c.Continue()
}
