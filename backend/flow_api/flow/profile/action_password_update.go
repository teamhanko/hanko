package profile

import (
	"fmt"

	auditlog "github.com/teamhanko/hanko/backend/v2/audit_log"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flow_api/services"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
	"github.com/teamhanko/hanko/backend/v2/webhooks/utils"
)

type PasswordUpdate struct {
	shared.Action
}

func (a PasswordUpdate) GetName() flowpilot.ActionName {
	return shared.ActionPasswordUpdate
}

func (a PasswordUpdate) GetDescription() string {
	return "Update an existing password."
}

func (a PasswordUpdate) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	userModel, _ := c.Get("session_user").(*models.User)

	if !deps.Cfg.Password.Enabled {
		c.SuspendAction()
	}

	if userModel.PasswordCredential == nil {
		// The password_create action must be used instead
		c.SuspendAction()
	}

	c.AddInputs(flowpilot.StringInput("password").
		Required(true).
		MinLength(deps.Cfg.Password.MinLength).
		MaxLength(72))
}

func (a PasswordUpdate) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	password := c.Input().Get("password").String()

	err := deps.PasswordService.UpdatePassword(deps.Tx, userModel.PasswordCredential, password)
	if err != nil {
		return fmt.Errorf("could not udate password: %w", err)
	}

	if deps.Cfg.SecurityNotifications.Notifications.PasswordUpdate.Enabled {
		deps.SecurityNotificationService.SendNotification(deps.Tx, services.SendSecurityNotificationParams{
			EmailAddress: userModel.Emails.GetPrimary().Address,
			Template:     "password_update",
			Language:     deps.HttpContext.Request().Header.Get("X-Language"),
		})
	}

	err = deps.AuditLogger.CreateWithConnection(
		deps.Tx,
		deps.HttpContext,
		models.AuditLogPasswordChanged,
		&models.User{ID: userModel.ID},
		nil,
		auditlog.Detail("context", "profile"),
		auditlog.Detail("flow_id", c.GetFlowID()))

	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	utils.NotifyUserChange(deps.HttpContext, deps.Tx, deps.Persister, events.UserPasswordChange, userModel.ID)

	return c.Continue(shared.StateProfileInit)
}
