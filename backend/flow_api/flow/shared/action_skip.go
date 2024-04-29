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

func (a Skip) Initialize(c flowpilot.InitializationContext) {
	//if !c.Stash().Get("skip_to").Exists() || c.Stash().Get("skip_to").String() == "" {
	//	c.SuspendAction()
	//}

	//if !c.Stash().Get("skip_from").Exists() || c.Stash().Get("skip_from").String() == "" {
	//	c.SuspendAction()
	//}

	//if nextState, ok := c.GetNextStateDuringInitForSchemaCreation(); ok {
	//	if nextState != flowpilot.StateName(c.Stash().Get("skip_from").String()) {
	//		c.SuspendAction()
	//	}
	//}

}

func (a Skip) Execute(c flowpilot.ExecutionContext) error {
	return c.EndSubFlow()
}

func (a Skip) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
