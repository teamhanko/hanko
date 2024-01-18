package profile

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type GetProfileData struct {
	shared.Action
}

func (h GetProfileData) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	sessionToken, ok := deps.HttpContext.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}

	userId, err := uuid.FromString(sessionToken.Subject())
	if err != nil {
		return fmt.Errorf("failed to parse userId from JWT subject: %w", err)
	}

	userModel, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(userId)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	if userModel == nil {
		return errors.New("user not found")
	}

	c.Set("session_user", userModel)

	err = c.Payload().Set("user", dto.ProfileDataFromUserModel(userModel))
	if err != nil {
		return fmt.Errorf("failed to set user payload: %w", err)
	}

	return nil
}
