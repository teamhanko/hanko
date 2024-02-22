package profile

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
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

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted)
	}

	passwordCredential, err := deps.Persister.GetPasswordCredentialPersisterWithConnection(deps.Tx).GetByUserID(userModel.ID)
	if err != nil {
		return fmt.Errorf("could not fetch password credential: %w", err)
	}

	password := c.Input().Get("password").String()

	if passwordCredential == nil {
		err = deps.PasswordService.CreatePassword(userModel.ID, password)
	} else {
		err = deps.PasswordService.UpdatePassword(passwordCredential, password)
	}

	if err != nil {
		return fmt.Errorf("could not set password: %w", err)
	}

	return c.ContinueFlow(StateProfileInit)
}

func (a PasswordSet) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
