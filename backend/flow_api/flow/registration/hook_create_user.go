package registration

import (
	"encoding/json"
	"fmt"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
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
	// Set by shared thirdparty_oauth action because the third party callback endpoint already
	// creates the user.
	if c.Stash().Get("skip_user_creation").Bool() {
		return nil
	}

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
		c,
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

func (h CreateUser) createUser(c flowpilot.HookExecutionContext, id uuid.UUID, email string, emailVerified bool, username string, passkey *models.WebauthnCredential, password string) error {
	deps := h.GetDeps(c)

	now := time.Now().UTC()

	var auditLogDetails []auditlog.DetailOption

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

		auditLogDetails = append(auditLogDetails, auditlog.Detail("passkey", passkey.ID))
	}

	if password != "" {
		err = deps.Persister.GetPasswordCredentialPersisterWithConnection(deps.Tx).Create(models.PasswordCredential{
			UserId:   id,
			Password: password,
		})
		if err != nil {
			return err
		}

		auditLogDetails = append(auditLogDetails, auditlog.Detail("password", true))
	}

	user, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(id)
	if err != nil {
		return err
	}

	if user.Username != "" {
		auditLogDetails = append(auditLogDetails, auditlog.Detail("username", user.Username))
	}

	auditLogDetails = append(auditLogDetails, auditlog.Detail("flow_id", c.GetFlowID()))

	err = deps.AuditLogger.Create(
		deps.HttpContext,
		models.AuditLogUserCreated,
		user,
		nil,
		auditLogDetails...,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w")
	}

	return nil
}
