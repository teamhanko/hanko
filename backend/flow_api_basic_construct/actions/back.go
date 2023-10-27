package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewBack() flowpilot.Action {
	return Back{}
}

type Back struct{}

func (a Back) GetName() flowpilot.ActionName {
	return common.ActionBack
}

func (a Back) GetDescription() string {
	return "Navigate one step back."
}

func (a Back) Initialize(_ flowpilot.InitializationContext) {}

func (a Back) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueToPreviousState()
}
