package shared

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type WebauthnCredentialSave struct {
	Action
}

func (h WebauthnCredentialSave) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	if !c.Stash().Get(StashPathUserID).Exists() {
		return nil
	}

	userId, err := uuid.FromString(c.Stash().Get(StashPathUserID).String())
	if err != nil {
		return fmt.Errorf("failed to parse stashed user_id into a uuid: %w", err)
	}

	if !c.Stash().Get(StashPathWebauthnCredentials).Exists() {
		return nil
	}

	auditLogDetails := []auditlog.DetailOption{
		auditlog.Detail("flow_id", c.GetFlowID()),
	}

	auditLogType := models.AuditLogPasskeyCreated

	for _, webauthnCredential := range c.Stash().Get(StashPathWebauthnCredentials).Array() {
		var credentialModel models.WebauthnCredential
		err = json.Unmarshal([]byte(webauthnCredential.String()), &credentialModel)
		if err != nil {
			return fmt.Errorf("failed to unmarshal stashed webauthn_credential: %w", err)
		}

		err = deps.Persister.GetWebauthnCredentialPersisterWithConnection(deps.Tx).Create(credentialModel)
		if err != nil {
			return err
		}

		if userModel, ok := c.Get("session_user").(*models.User); ok {
			userModel.WebauthnCredentials = append(userModel.WebauthnCredentials, credentialModel)
		}

		if credentialModel.MFAOnly {
			auditLogType = models.AuditLogSecurityKeyCreated
			auditLogDetails = append(auditLogDetails, auditlog.Detail("security_key", credentialModel.ID))
		} else {
			auditLogDetails = append(auditLogDetails, auditlog.Detail("passkey", credentialModel.ID))
		}
	}

	err = c.Stash().Delete(StashPathWebauthnCredentials)
	if err != nil {
		return err
	}

	err = deps.AuditLogger.CreateWithConnection(
		deps.Tx,
		deps.HttpContext,
		auditLogType,
		&models.User{ID: userId},
		nil,
		auditLogDetails...)

	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	return nil
}
