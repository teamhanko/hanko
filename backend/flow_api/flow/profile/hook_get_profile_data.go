package profile

import (
	"errors"
	"fmt"

	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v3/flowpilot"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type GetProfileData struct {
	shared.Action
}

func (h GetProfileData) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return errors.New("no valid session")
	}

	profileData := dto.ProfileDataFromUserModel(userModel, &deps.Cfg.TenantConfig)

	err := c.Payload().Set("user", profileData)
	if err != nil {
		return fmt.Errorf("failed to set user payload: %w", err)
	}

	return nil
}
