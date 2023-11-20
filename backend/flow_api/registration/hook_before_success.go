package registration

import (
	"encoding/json"
	"fmt"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/dto/intern"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type BeforeSuccess struct {
	shared.Action
}

func (h BeforeSuccess) Execute(c flowpilot.HookExecutionContext) error {
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

	passkeyCredentialStr := c.Stash().Get("passkey_credential").String()
	var passkeyCredential webauthnLib.Credential
	err = json.Unmarshal([]byte(passkeyCredentialStr), &passkeyCredential)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stashed passkey_credential: %w", err)
	}
	passkeyBackupEligible := c.Stash().Get("passkey_backup_eligible").Bool()
	passkeyBackupState := c.Stash().Get("passkey_backup_state").Bool()

	credentialModel := intern.WebauthnCredentialToModel(&passkeyCredential, userId, passkeyBackupEligible, passkeyBackupState)
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

	sessionToken, err := deps.SessionManager.GenerateJWT(userId)
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %w", err)
	}
	cookie, err := deps.SessionManager.GenerateCookie(sessionToken)
	if err != nil {
		return fmt.Errorf("failed to generate auth cookie, %w", err)
	}

	deps.HttpContext.SetCookie(cookie)

	return nil
}

func (h BeforeSuccess) createUser(deps *shared.Dependencies, id uuid.UUID, email string, emailVerified bool, username string, passkey *models.WebauthnCredential, password string) error {
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
