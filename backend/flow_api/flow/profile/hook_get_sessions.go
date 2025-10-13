package profile

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type GetSessions struct {
	shared.Action
}

func (h GetSessions) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	if !deps.Cfg.Session.ShowOnProfile {
		return nil
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return errors.New("no valid session")
	}

	activeSessions, err := deps.Persister.GetSessionPersisterWithConnection(deps.Tx).ListActive(userModel.ID)
	if err != nil {
		return fmt.Errorf("failed to get sessions from db: %w", err)
	}

	currentSessionID := uuid.FromStringOrNil(c.Get("session_id").(string))

	sessionsDto := make([]dto.SessionData, len(activeSessions))
	for i := range activeSessions {
		sessionsDto[i] = dto.FromSessionModel(activeSessions[i], activeSessions[i].ID == currentSessionID)
	}

	err = c.Payload().Set("sessions", sessionsDto)
	if err != nil {
		return fmt.Errorf("failed to set sessions payload: %w", err)
	}

	return nil
}
