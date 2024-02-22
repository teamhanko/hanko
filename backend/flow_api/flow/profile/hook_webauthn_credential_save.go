package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"github.com/teamhanko/hanko/backend/dto/intern"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type WebauthnCredentialSave struct {
	shared.Action
}

func (h WebauthnCredentialSave) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return flowpilot.ErrorOperationNotPermitted
	}

	if !c.Stash().Get("webauthn_credential").Exists() {
		return errors.New("webauthn_credential not set in stash")
	}

	webauthnCredentialJson := c.Stash().Get("webauthn_credential").String()

	var webauthnCredential webauthnLib.Credential
	err := json.Unmarshal([]byte(webauthnCredentialJson), &webauthnCredential)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stashed webauthn_credential: %w", err)
	}

	credentialModel := intern.WebauthnCredentialToModel(&webauthnCredential, userModel.ID, webauthnCredential.Flags.BackupEligible, webauthnCredential.Flags.BackupState, deps.AuthenticatorMetadata)
	err = deps.Persister.GetWebauthnCredentialPersisterWithConnection(deps.Tx).Create(*credentialModel)
	if err != nil {
		return fmt.Errorf("failed so save credential: %w", err)
	}
	return nil
}
