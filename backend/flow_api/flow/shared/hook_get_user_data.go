package shared

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type GetUserData struct {
	Action
}

func (h GetUserData) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	userId, err := uuid.FromString(c.Stash().Get("user_id").String())
	if err != nil {
		return fmt.Errorf("failed to parse stashed user_id into a uuid: %w", err)
	}

	userModel, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user from db: %w", err)
	}

	err = c.Payload().Set("user", dto.ProfileDataFromUserModel(userModel))
	if err != nil {
		return fmt.Errorf("failed to set user payload: %w", err)
	}

	return nil
}
