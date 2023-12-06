package profile

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
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

	if !c.Stash().Get("user_id").Exists() {
		return c.ContinueFlowWithError(
			c.GetErrorState(),
			flowpilot.ErrorOperationNotPermitted.
				Wrap(errors.New("user_id does not exist")))
	}

	userId := c.Stash().Get("user_id").String()

	userModel, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(uuid.FromStringOrNil(userId))
	if err != nil {
		return fmt.Errorf("could not fetch user: %w", err)
	}

	if userModel == nil {
		return errors.New("user not found")
	}

	username := c.Input().Get("username").String()
	userModel.Username = username

	err = deps.Persister.GetUserPersisterWithConnection(deps.Tx).Update(*userModel)
	if err != nil {
		return fmt.Errorf("could not update user: %w", err)
	}

	return c.ContinueFlow(StateProfileInit)
}
