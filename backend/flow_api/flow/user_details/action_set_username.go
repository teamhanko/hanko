package user_details

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
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
		MaxLength(deps.Cfg.Username.MaxLength))
}

func (a UsernameSet) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	userID := uuid.FromStringOrNil(c.Stash().Get("user_id").String())
	user, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(userID)
	if err != nil {
		return fmt.Errorf("failed to get user from db: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user does not exists (id: %s)", userID.String())
	}

	user.Username = c.Input().Get("username").String()

	duplicateUser, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).GetByUsername(user.Username)
	if err != nil {
		return fmt.Errorf("failed to get user from db: %w", err)
	}

	if duplicateUser != nil && duplicateUser.ID.String() != user.ID.String() {
		c.Input().SetError("username", shared.ErrorUsernameAlreadyExists)
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	err = deps.Persister.GetUserPersisterWithConnection(deps.Tx).Update(*user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return c.EndSubFlow()
}

func (a UsernameSet) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
