package profile

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type EmailVerify struct {
	shared.Action
}

func (a EmailVerify) GetName() flowpilot.ActionName {
	return shared.ActionEmailVerify
}

func (a EmailVerify) GetDescription() string {
	return "Verify an email."
}

func (a EmailVerify) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Identifier.Email.Enabled || !deps.Cfg.Passcode.Enabled {
		c.SuspendAction()
		return
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		c.SuspendAction()
		return
	}

	if !userModel.Emails.HasUnverified() {
		c.SuspendAction()
		return
	}

	c.AddInputs(flowpilot.StringInput("email_id").Required(true).Hidden(true))
}

func (a EmailVerify) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted)
	}

	emailModel := userModel.GetEmailById(uuid.FromStringOrNil(c.Input().Get("email_id").String()))
	if emailModel == nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), shared.ErrorNotFound)
	}

	err := c.Stash().Set("email", emailModel.Address)
	if err != nil {
		return fmt.Errorf("failed to set email address to verify to stash: %w", err)
	}

	err = c.Stash().Set("user_id", userModel.ID.String())
	if err != nil {
		return fmt.Errorf("failed to set user_id to stash: %w", err)
	}

	err = c.Stash().Set("passcode_template", "email_verification")
	if err != nil {
		return fmt.Errorf("failed to set passcode_tempalte to stash %w", err)
	}

	return c.StartSubFlow(shared.StatePasscodeConfirmation, shared.StateProfileInit)
}

func (a EmailVerify) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
