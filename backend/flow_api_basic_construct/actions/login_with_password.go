package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewLoginWithPassword() flowpilot.Action {
	return LoginWithPassword{}
}

type LoginWithPassword struct{}

func (a LoginWithPassword) GetName() flowpilot.ActionName {
	return common.ActionLoginWithPassword
}

func (a LoginWithPassword) GetDescription() string {
	return "Login with a password."
}

func (a LoginWithPassword) Initialize(_ flowpilot.InitializationContext) {}

func (a LoginWithPassword) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueFlow(common.StatePasswordLogin)
}
