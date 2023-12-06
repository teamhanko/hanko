package profile

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type EmailVerify struct {
	shared.Action
}

func (a EmailVerify) GetName() flowpilot.ActionName {
	return ActionEmailVerify
}

func (a EmailVerify) GetDescription() string {
	return "Verify an email."
}

func (a EmailVerify) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("email_id").Required(true).Hidden(true))
}

func (a EmailVerify) Execute(c flowpilot.ExecutionContext) error {
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

	userModel, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(userId)
	if err != nil {
		return fmt.Errorf("could not fetch user: %w", err)
	}

	emailModel := userModel.GetEmailById(uuid.FromStringOrNil(c.Input().Get("email_id").String()))
	if emailModel == nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), shared.ErrorNotFound)
	}

	err = c.Stash().Set("email", emailModel.Address)
	if err != nil {
		return fmt.Errorf("failed to set email address to verify to stash: %w", err)
	}

	err = c.Stash().Set("passcode_template", "email_verification")
	if err != nil {
		return fmt.Errorf("failed to set passcode_tempalte to stash %w", err)
	}

	return c.StartSubFlow(passcode.StatePasscodeConfirmation, StateProfileInit)
}
