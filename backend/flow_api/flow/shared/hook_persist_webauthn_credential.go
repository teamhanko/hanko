package shared

import (
	"encoding/json"
	"fmt"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/dto/intern"
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

	if !c.Stash().Get(StashPathWebauthnCredential).Exists() {
		return nil
	}

	webauthnCredentialJson := c.Stash().Get(StashPathWebauthnCredential).String()

	var webauthnCredential webauthnLib.Credential
	err = json.Unmarshal([]byte(webauthnCredentialJson), &webauthnCredential)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stashed webauthn_credential: %w", err)
	}

	var mfaOnly bool
	auditLogType := models.AuditLogPasskeyCreated
	if c.Stash().Get(StashPathCreateMFAOnlyCredential).Bool() {
		mfaOnly = true
		auditLogType = models.AuditLogSecurityKeyCreated
	}

	credentialModel := intern.WebauthnCredentialToModel(&webauthnCredential, userId, webauthnCredential.Flags.BackupEligible, webauthnCredential.Flags.BackupState, mfaOnly, deps.AuthenticatorMetadata)
	err = deps.Persister.GetWebauthnCredentialPersisterWithConnection(deps.Tx).Create(*credentialModel)
	if err != nil {
		return fmt.Errorf("failed so save credential: %w", err)
	}

	auditLogDetails := []auditlog.DetailOption{
		auditlog.Detail("flow_id", c.GetFlowID()),
	}

	if credentialModel.MFAOnly {
		auditLogDetails = append(auditLogDetails, auditlog.Detail("security_key", credentialModel.ID))
	} else {
		auditLogDetails = append(auditLogDetails, auditlog.Detail("passkey", credentialModel.ID))
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

	if userModel, ok := c.Get("session_user").(*models.User); ok {
		userModel.WebauthnCredentials = append(userModel.WebauthnCredentials, *credentialModel)
	}

	return nil
}
