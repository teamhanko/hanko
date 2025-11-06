package profile

import (
	"fmt"

	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/v2/audit_log"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flow_api/services"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
	"github.com/teamhanko/hanko/backend/v2/webhooks/utils"
)

type EmailSetPrimary struct {
	shared.Action
}

func (a EmailSetPrimary) GetName() flowpilot.ActionName {
	return shared.ActionEmailSetPrimary
}

func (a EmailSetPrimary) GetDescription() string {
	return "Sets a an email address as the primary email address."
}

func (a EmailSetPrimary) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Email.Enabled {
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
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	emailId := uuid.FromStringOrNil(c.Input().Get("email_id").String())
	emailModel := userModel.GetEmailById(emailId)

	if emailModel == nil {
		return c.Error(shared.ErrorNotFound)
	}

	if emailModel.IsPrimary() {
		return c.Continue(shared.StateProfileInit)
	}

	var primaryEmail *models.PrimaryEmail
	var existingPrimaryEmailAddress string
	if e := userModel.Emails.GetPrimary(); e != nil {
		primaryEmail = e.PrimaryEmail
		existingPrimaryEmailAddress = e.Address
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

	if deps.Cfg.SecurityNotifications.Notifications.EmailCreate.Enabled && existingPrimaryEmailAddress != "" {
		deps.SecurityNotificationService.SendNotification(deps.Tx, services.SendSecurityNotificationParams{
			EmailAddress: existingPrimaryEmailAddress,
			Template:     "primary_email_update",
			HttpContext:  deps.HttpContext,
			BodyData: map[string]interface{}{
				"OldEmailAddress": existingPrimaryEmailAddress,
				"NewEmailAddress": emailModel.Address,
			},
		})
	}

	err := deps.AuditLogger.CreateWithConnection(
		deps.Tx,
		deps.HttpContext,
		models.AuditLogPrimaryEmailChanged,
		&models.User{ID: userModel.ID},
		nil,
		auditlog.Detail("email", emailModel.Address),
		auditlog.Detail("flow_id", c.GetFlowID()))

	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	userModel.SetPrimaryEmail(primaryEmail)
	utils.NotifyUserChange(deps.HttpContext, deps.Tx, deps.Persister, events.UserEmailPrimary, userModel.ID)

	return c.Continue(shared.StateProfileInit)
}
