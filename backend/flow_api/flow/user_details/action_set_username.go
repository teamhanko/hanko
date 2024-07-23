package user_details

import (
	"fmt"
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type UsernameSet struct {
	shared.Action
}

func (a UsernameSet) GetName() flowpilot.ActionName {
	return shared.ActionUsernameSet
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
	user, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(userID)
	if err != nil {
		return fmt.Errorf("failed to get user from db: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user does not exists (id: %s)", userID.String())
	}

	username := c.Input().Get("username").String()

	if !services.ValidateUsername(username) {
		c.Input().SetError("username", shared.ErrorInvalidUsername)
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	user.Username = nulls.NewString(username)
	duplicateUser, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).GetByUsername(user.Username.String)
	if err != nil {
		return fmt.Errorf("failed to get user from db: %w", err)
	}

	if duplicateUser != nil && duplicateUser.ID.String() != user.ID.String() {
		c.Input().SetError("username", shared.ErrorUsernameAlreadyExists)
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	err = deps.Persister.GetUserPersisterWithConnection(deps.Tx).Update(*user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	c.PreventRevert()

	return c.Continue()
}
