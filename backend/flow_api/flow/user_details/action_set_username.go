package user_details

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type UsernameSet struct {
	shared.Action
}

func (a UsernameSet) GetName() flowpilot.ActionName {
	return shared.ActionUsernameCreate
}

func (a UsernameSet) GetDescription() string {
	return "Set a new username."
}

func (a UsernameSet) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	c.AddInputs(flowpilot.StringInput("username").
		Required(!deps.Cfg.Username.Optional).
		MinLength(deps.Cfg.Username.MinLength).
		MaxLength(deps.Cfg.Username.MaxLength).
		TrimSpace(true).
		LowerCase(true))
}

func (a UsernameSet) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	userID := uuid.FromStringOrNil(c.Stash().Get(shared.StashPathUserID).String())
	username := c.Input().Get("username").String()

	if !services.ValidateUsername(username) {
		c.Input().SetError("username", shared.ErrorInvalidUsername)
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	duplicateUsername, err := deps.Persister.GetUsernamePersisterWithConnection(deps.Tx).GetByName(username)
	if err != nil {
		return fmt.Errorf("failed to get user from db: %w", err)
	}

	if duplicateUsername != nil && duplicateUsername.ID.String() != userID.String() {
		c.Input().SetError("username", shared.ErrorUsernameAlreadyExists)
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	usernameModel := models.NewUsername(userID, username)
	err = deps.Persister.GetUsernamePersisterWithConnection(deps.Tx).Create(*usernameModel)
	if err != nil {
		return fmt.Errorf("failed to create username: %w", err)
	}

	c.PreventRevert()

	return c.Continue()
}
