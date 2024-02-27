package registration

import (
	"encoding/json"
	"fmt"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/dto/intern"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type CreateUser struct {
	shared.Action
}

func (h CreateUser) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	userId, err := uuid.NewV4()
	if err != nil {
		return err
	}
	if c.Stash().Get("user_id").Exists() {
		userId, err = uuid.FromString(c.Stash().Get("user_id").String())
		if err != nil {
			return fmt.Errorf("failed to parse stashed user_id into a uuid: %w", err)
		}
	}

	var credentialModel *models.WebauthnCredential
	if c.Stash().Get("webauthn_credential").Exists() {
		webauthnCredentialStr := c.Stash().Get("webauthn_credential").String()

		var webauthnCredential webauthnLib.Credential
		err = json.Unmarshal([]byte(webauthnCredentialStr), &webauthnCredential)
		if err != nil {
			return fmt.Errorf("failed to unmarshal stashed webauthn_credential: %w", err)
		}

		// TODO: Who/what sets this? Do we need this?
		passkeyBackupEligible := c.Stash().Get("passkey_backup_eligible").Bool()
		passkeyBackupState := c.Stash().Get("passkey_backup_state").Bool()

		credentialModel = intern.WebauthnCredentialToModel(&webauthnCredential, userId, passkeyBackupEligible, passkeyBackupState, deps.AuthenticatorMetadata)
	}

	err = h.createUser(
		deps,
		userId,
		c.Stash().Get("email").String(),
		c.Stash().Get("email_verified").Bool(),
		c.Stash().Get("username").String(),
		credentialModel,
		c.Stash().Get("new_password").String(),
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (h CreateUser) createUser(deps *shared.Dependencies, id uuid.UUID, email string, emailVerified bool, username string, passkey *models.WebauthnCredential, password string) error {
	// TODO: add audit log
	now := time.Now().UTC()
	err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Create(models.User{
		ID:        id,
		Username:  username,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return err
	}

	if email != "" {
		emailModel := models.NewEmail(&id, email)
		emailModel.Verified = emailVerified
		err = deps.Persister.GetEmailPersisterWithConnection(deps.Tx).Create(*emailModel)
		if err != nil {
			return err
		}

		primaryEmail := models.NewPrimaryEmail(emailModel.ID, id)
		err = deps.Persister.GetPrimaryEmailPersisterWithConnection(deps.Tx).Create(*primaryEmail)
		if err != nil {
			return err
		}
	}

	if passkey != nil {
		err = deps.Persister.GetWebauthnCredentialPersisterWithConnection(deps.Tx).Create(*passkey)
		if err != nil {
			return err
		}
	}

	if password != "" {
		err = deps.Persister.GetPasswordCredentialPersisterWithConnection(deps.Tx).Create(models.PasswordCredential{
			UserId:   id,
			Password: password,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
