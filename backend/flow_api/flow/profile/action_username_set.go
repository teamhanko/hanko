package profile

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type UsernameSet struct {
	shared.Action
}

func (a UsernameSet) GetName() flowpilot.ActionName {
	return ActionUsernameSet
}

func (a UsernameSet) GetDescription() string {
	return "Set the username of a user."
}

func (a UsernameSet) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Identifier.Username.Enabled {
		c.SuspendAction()
	} else {
		c.AddInputs(flowpilot.StringInput("username").Required(true))
	}
}

func (a UsernameSet) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted)
	}

	username := c.Input().Get("username").String()
	userModel.Username = username

	err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Update(*userModel)
	if err != nil {
		return fmt.Errorf("could not update user: %w", err)
	}

	return c.ContinueFlow(StateProfileInit)
}
