package profile

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
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

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted)
	}

	err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Delete(*userModel)
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
