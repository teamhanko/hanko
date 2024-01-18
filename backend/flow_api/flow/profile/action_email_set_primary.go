package profile

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type EmailSetPrimary struct {
	shared.Action
}

func (a EmailSetPrimary) GetName() flowpilot.ActionName {
	return ActionEmailSetPrimary
}

func (a EmailSetPrimary) GetDescription() string {
	return "Sets a an email address as the primary email address."
}

func (a EmailSetPrimary) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Identifier.Email.Enabled {
		c.SuspendAction()
		return
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		c.SuspendAction()
		return
	}

	if len(userModel.Emails) == 1 && userModel.Emails[0].IsPrimary() {
		c.SuspendAction()
		return
	}

	if len(userModel.Emails) == 0 {
		c.SuspendAction()
		return
	}

	c.AddInputs(flowpilot.StringInput("email_id").Required(true).Hidden(true))
}

func (a EmailSetPrimary) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted)
	}

	emailId := uuid.FromStringOrNil(c.Input().Get("email_id").String())
	emailModel := userModel.GetEmailById(emailId)

	if emailModel == nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), shared.ErrorNotFound)
	}

	if emailModel.IsPrimary() {
		return c.ContinueFlow(StateProfileInit)
	}

	var primaryEmail *models.PrimaryEmail
	if e := userModel.Emails.GetPrimary(); e != nil {
		primaryEmail = e.PrimaryEmail
	}

	if primaryEmail == nil {
		primaryEmail = models.NewPrimaryEmail(emailModel.ID, userModel.ID)
		err := deps.Persister.GetPrimaryEmailPersisterWithConnection(deps.Tx).Create(*primaryEmail)
		if err != nil {
			return fmt.Errorf("failed to store new primary email: %w", err)
		}
	} else {
		primaryEmail.EmailID = emailModel.ID
		err := deps.Persister.GetPrimaryEmailPersisterWithConnection(deps.Tx).Update(*primaryEmail)
		if err != nil {
			return fmt.Errorf("failed to change primary email: %w", err)
		}
	}

	return c.ContinueFlow(StateProfileInit)
}
