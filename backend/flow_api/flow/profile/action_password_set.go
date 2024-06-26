package profile

import (
	"fmt"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type PasswordSet struct {
	shared.Action
}

func (a PasswordSet) GetName() flowpilot.ActionName {
	return shared.ActionPasswordSet
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
			MinLength(deps.Cfg.Password.MinLength).
			MaxLength(72).
			Persist(false),
		)
	}
}

func (a PasswordSet) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	passwordCredential, err := deps.Persister.GetPasswordCredentialPersisterWithConnection(deps.Tx).GetByUserID(userModel.ID)
	if err != nil {
		return fmt.Errorf("could not fetch password credential: %w", err)
	}

	password := c.Input().Get("password").String()

	if passwordCredential == nil {
		passwordCredential = models.NewPasswordCredential(userModel.ID, "") // ?
		err = deps.PasswordService.CreatePassword(userModel.ID, password)
	} else {
		err = deps.PasswordService.UpdatePassword(passwordCredential, password)
	}

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
