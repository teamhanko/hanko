package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/dto/intern"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type WebauthnCredentialSave struct {
	shared.Action
}

func (h WebauthnCredentialSave) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	if !c.Stash().Get("passkey_credential").Exists() {
		return errors.New("passkey_credential not set in stash")
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

	credentialModel := intern.WebauthnCredentialToModel(&webauthnCredential, userModel.ID, webauthnCredential.Flags.BackupEligible, webauthnCredential.Flags.BackupState)
	err = deps.Persister.GetWebauthnCredentialPersisterWithConnection(deps.Tx).Create(*credentialModel)
	if err != nil {
		return fmt.Errorf("failed so save credential: %w", err)
	}
	return nil
}
