package shared

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type Back struct{}

func (a Back) GetName() flowpilot.ActionName {
	return ActionBack
}

func (a Back) GetDescription() string {
	return "Navigate one step back."
}

func (a Back) Initialize(_ flowpilot.InitializationContext) {}

func (a Back) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueToPreviousState()
}

func (a Back) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
