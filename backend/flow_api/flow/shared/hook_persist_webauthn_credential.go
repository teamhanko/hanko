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

	if !c.Stash().Get("user_id").Exists() {
		return nil
	}

	userId, err := uuid.FromString(c.Stash().Get("user_id").String())
	if err != nil {
		return fmt.Errorf("failed to parse stashed user_id into a uuid: %w", err)
	}

	if !c.Stash().Get("webauthn_credential").Exists() {
		return nil
	}

	webauthnCredentialJson := c.Stash().Get("webauthn_credential").String()

	var webauthnCredential webauthnLib.Credential
	err = json.Unmarshal([]byte(webauthnCredentialJson), &webauthnCredential)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stashed webauthn_credential: %w", err)
	}

	credentialModel := intern.WebauthnCredentialToModel(&webauthnCredential, userId, webauthnCredential.Flags.BackupEligible, webauthnCredential.Flags.BackupState, deps.AuthenticatorMetadata)
	err = deps.Persister.GetWebauthnCredentialPersisterWithConnection(deps.Tx).Create(*credentialModel)
	if err != nil {
		return fmt.Errorf("failed so save credential: %w", err)
	}

	err = deps.AuditLogger.CreateWithConnection(
		deps.Tx,
		deps.HttpContext,
		models.AuditLogPasskeyCreated,
		&models.User{ID: userId},
		nil,
		auditlog.Detail("credential_id", credentialModel.ID),
		auditlog.Detail("flow_id", c.GetFlowID()))

	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	if userModel, ok := c.Get("session_user").(*models.User); ok {
		userModel.WebauthnCredentials = append(userModel.WebauthnCredentials, *credentialModel)
	}

	return nil
}
