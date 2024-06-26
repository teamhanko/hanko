package credential_usage

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type PasswordRecovery struct {
	shared.Action
}

func (a PasswordRecovery) GetName() flowpilot.ActionName {
	return shared.ActionPasswordRecovery
}

func (a PasswordRecovery) GetDescription() string {
	return "Submit a new password."
}

func (a PasswordRecovery) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	c.AddInputs(flowpilot.PasswordInput("new_password").
		Required(true).
		MinLength(deps.Cfg.Password.MinLength),
	)

	if !deps.Cfg.Password.Enabled {
		c.SuspendAction()
	}
}

func (a PasswordRecovery) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	newPassword := c.Input().Get("new_password").String()

	if !c.Stash().Get(shared.StashPathUserID).Exists() {
		return c.Error(flowpilot.ErrorOperationNotPermitted.Wrap(errors.New("user_id does not exist")))
	}

	authUserID := c.Stash().Get(shared.StashPathUserID).String()

	err := deps.PasswordService.RecoverPassword(uuid.FromStringOrNil(authUserID), newPassword)

	if err != nil {
		if errors.Is(err, services.ErrorPasswordInvalid) {
			c.Input().SetError("password", flowpilot.ErrorValueInvalid)
			return c.Error(flowpilot.ErrorFormDataInvalid.Wrap(err))
		}

		return fmt.Errorf("could not recover password: %w", err)
	}

	err = deps.AuditLogger.CreateWithConnection(
		deps.Tx,
		deps.HttpContext,
		models.AuditLogPasswordChanged,
		&models.User{ID: uuid.FromStringOrNil(authUserID)},
		nil,
		auditlog.Detail("context", "recovery"),
		auditlog.Detail("flow_id", c.GetFlowID()))

	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	err = c.Stash().Set(shared.StashPathUserHasPassword, true)
	if err != nil {
		return fmt.Errorf("failed to set user_has_password to the stash: %w", err)
	}

	err = c.DeleteStateHistory(true)
	if err != nil {
		return fmt.Errorf("failed to delete the state history: %w", err)
	}

	return c.Continue()
}
