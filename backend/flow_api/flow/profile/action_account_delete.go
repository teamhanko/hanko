package profile

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type AccountDelete struct {
	shared.Action
}

func (a AccountDelete) GetName() flowpilot.ActionName {
	return ActionAccountDelete
}

func (a AccountDelete) GetDescription() string {
	return "Delete an account."
}

func (a AccountDelete) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Account.AllowDeletion {
		c.SuspendAction()
	}
}

func (a AccountDelete) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if !c.Stash().Get("user_id").Exists() {
		return c.ContinueFlowWithError(
			c.GetErrorState(),
			flowpilot.ErrorOperationNotPermitted.
				Wrap(errors.New("user_id does not exist")))
	}

	userId := uuid.FromStringOrNil(c.Stash().Get("user_id").String())

	userModel, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(userId)
	if err != nil {
		return fmt.Errorf("could not fetch user: %w", err)
	}

	if userModel == nil {
		return errors.New("user not found")
	}

	err = deps.Persister.GetUserPersisterWithConnection(deps.Tx).Delete(*userModel)
	if err != nil {
		return fmt.Errorf("could not delete user: %w", err)
	}

	cookie, err := deps.SessionManager.DeleteCookie()
	if err != nil {
		return fmt.Errorf("could not delete cookie: %w", err)
	}

	deps.HttpContext.SetCookie(cookie)

	return c.ContinueFlow(StateProfileAccountDeleted)
}
