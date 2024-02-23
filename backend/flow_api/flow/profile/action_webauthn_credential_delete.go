package profile

import (
	"fmt"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type WebauthnCredentialDelete struct {
	shared.Action
}

func (a WebauthnCredentialDelete) GetName() flowpilot.ActionName {
	return ActionWebauthnCredentialDelete
}

func (a WebauthnCredentialDelete) GetDescription() string {
	return "Delete a Webauthn credential."
}

func (a WebauthnCredentialDelete) Initialize(c flowpilot.InitializationContext) {
	if a.mustSuspend(c) {
		c.SuspendAction()
		return
	}

	c.AddInputs(flowpilot.StringInput("passkey_id").Required(true).Hidden(true))
}

func (a WebauthnCredentialDelete) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted)
	}

	webauthnCredentialModel := userModel.GetWebauthnCredentialById(c.Input().Get("passkey_id").String())
	if webauthnCredentialModel == nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), shared.ErrorNotFound)
	}

	err := deps.Persister.GetWebauthnCredentialPersisterWithConnection(deps.Tx).Delete(*webauthnCredentialModel)
	if err != nil {
		return fmt.Errorf("could not delete passkey: %w", err)
	}

	err = deps.AuditLogger.CreateWithConnection(
		deps.Tx,
		deps.HttpContext,
		models.AuditLogPasskeyDeleted,
		&models.User{ID: userModel.ID},
		nil,
		auditlog.Detail("credential_id", webauthnCredentialModel.ID),
		auditlog.Detail("flow_id", c.GetFlowID()))

	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	return c.ContinueFlow(StateProfileInit)
}

func (a WebauthnCredentialDelete) Finalize(c flowpilot.FinalizationContext) error {
	if a.mustSuspend(c) {
		c.SuspendAction()
	}

	return nil
}

func (a WebauthnCredentialDelete) mustSuspend(c flowpilot.Context) bool {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return true
	}

	if len(userModel.WebauthnCredentials) == 1 {
		if deps.Cfg.Passcode.Enabled && !deps.Cfg.Password.Enabled {
			if deps.Cfg.Identifier.Email.Optional && len(userModel.Emails) == 0 {
				return true
			}
		} else if !deps.Cfg.Passcode.Enabled && deps.Cfg.Password.Enabled {
			if userModel.PasswordCredential == nil {
				return true
			}
		} else {
			if len(userModel.Emails) == 0 && userModel.PasswordCredential == nil {
				return true
			}
		}
	}

	if len(userModel.WebauthnCredentials) == 0 {
		return true
	}
	return false
}
