package login_method_chooser

import (
	"github.com/teamhanko/hanko/backend/flow_api/constants"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPasswordLogin struct {
	shared.Action
}

func (a ContinueToPasswordLogin) GetName() flowpilot.ActionName {
	return constants.ActionContinueToPasswordLogin
}

func (a ContinueToPasswordLogin) GetDescription() string {
	return "Continue to the password login."
}

func (a ContinueToPasswordLogin) Initialize(c flowpilot.InitializationContext) {}

func (a ContinueToPasswordLogin) Execute(c flowpilot.ExecutionContext) error {
	return c.StartSubFlow(constants.StateLoginPassword)
}

func (a ContinueToPasswordLogin) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
