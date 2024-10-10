package profile

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type SessionDelete struct {
	shared.Action
}

func (a SessionDelete) GetName() flowpilot.ActionName {
	return shared.ActionSessionDelete
}

func (a SessionDelete) GetDescription() string {
	return "Delete a session."
}

func (a SessionDelete) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)
	if !deps.Cfg.Session.ServerSide.Enabled {
		c.SuspendAction()
	}
	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		c.SuspendAction()
		return
	}

	input := flowpilot.StringInput("session_id").Required(true).Hidden(true)

	currentSessionID := uuid.FromStringOrNil(c.Get("session_id").(string))
	sessions, err := deps.Persister.GetSessionPersister(deps.Tx).ListActive(userModel.ID)
	if err != nil {
		c.SuspendAction()
		return
	}

	for _, session := range sessions {
		if session.ID != currentSessionID {
			input.AllowedValue(session.ID.String(), session.ID.String())
		}
	}

	c.AddInputs(input)
}

func (a SessionDelete) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	sessionToBeDeleted := uuid.FromStringOrNil(c.Input().Get("session_id").String())

	session, err := deps.Persister.GetSessionPersister(deps.Tx).Get(sessionToBeDeleted)
	if err != nil {
		return fmt.Errorf("failed to get session from db: %w", err)
	}

	if session != nil {
		err = deps.Persister.GetSessionPersister(deps.Tx).Delete(*session)
	}

	return c.Continue(shared.StateProfileInit)
}
