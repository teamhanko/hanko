package login

import (
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type LoginWithPassword struct{}

func (a LoginWithPassword) GetName() flowpilot.ActionName {
	return shared.ActionLoginWithPassword
}

func (a LoginWithPassword) GetDescription() string {
	return "Login with a password."
}

func (a LoginWithPassword) Initialize(_ flowpilot.InitializationContext) {}

func (a LoginWithPassword) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueFlow(StateLoginPassword)
}
