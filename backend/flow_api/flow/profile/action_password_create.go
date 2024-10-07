package profile

import (
	"fmt"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type PasswordCreate struct {
	shared.Action
}

func (a PasswordCreate) GetName() flowpilot.ActionName {
	return shared.ActionPasswordCreate
}

func (a PasswordCreate) GetDescription() string {
	return "Create a new password."
}

func (a PasswordCreate) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	userModel, _ := c.Get("session_user").(*models.User)

	if !deps.Cfg.Password.Enabled {
		c.SuspendAction()
	}

	if userModel.PasswordCredential != nil {
		// The password_update action must be used instead
		c.SuspendAction()
	}

	c.AddInputs(flowpilot.StringInput("password").
		Required(true).
		MinLength(deps.Cfg.Password.MinLength).
		MaxLength(72),
	)

}

func (a PasswordCreate) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	password := c.Input().Get("password").String()

	passwordCredential := models.NewPasswordCredential(userModel.ID, password) // ?

	err := deps.PasswordService.CreatePassword(deps.Tx, userModel.ID, password) // ?
	if err != nil {
		return fmt.Errorf("could not set password: %w", err)
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

	userModel.PasswordCredential = passwordCredential

	return c.Continue(shared.StateProfileInit)
}
