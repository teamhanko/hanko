package profile

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type PasswordSet struct {
	shared.Action
}

func (a PasswordSet) GetName() flowpilot.ActionName {
	return ActionPasswordSet
}

func (a PasswordSet) GetDescription() string {
	return "Set a password."
}

func (a PasswordSet) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Password.Enabled {
		c.SuspendAction()
	} else {
		c.AddInputs(flowpilot.StringInput("password").
			Required(true).
			MinLength(deps.Cfg.Password.MinPasswordLength).
			Persist(false),
		)
	}
}

func (a PasswordSet) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	if !c.Stash().Get("user_id").Exists() {
		return c.ContinueFlowWithError(
			c.GetErrorState(),
			flowpilot.ErrorOperationNotPermitted.
				Wrap(errors.New("user_id does not exist")))
	}

	userId := uuid.FromStringOrNil(c.Stash().Get("user_id").String())

	passwordCredential, err := deps.Persister.GetPasswordCredentialPersisterWithConnection(deps.Tx).GetByUserID(userId)
	if err != nil {
		return fmt.Errorf("could not fetch password credential: %w", err)
	}

	password := c.Input().Get("password").String()

	if passwordCredential == nil {
		err = deps.PasswordService.CreatePassword(userId, password)
	} else {
		err = deps.PasswordService.UpdatePassword(passwordCredential, password)
	}

	if err != nil {
		return fmt.Errorf("could not set password: %w", err)
	}

	return c.ContinueFlow(StateProfileInit)
}
