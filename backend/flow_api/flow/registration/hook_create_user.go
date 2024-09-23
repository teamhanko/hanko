package registration

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"github.com/teamhanko/hanko/backend/webhooks/utils"
	"github.com/tidwall/gjson"
	"time"
)

type CreateUser struct {
	shared.Action
}

func (h CreateUser) Execute(c flowpilot.HookExecutionContext) error {
	// Set by shared thirdparty_oauth action because the third party callback endpoint already
	// creates the user.
	if c.Stash().Get(shared.StashPathSkipUserCreation).Bool() {
		return nil
	}

	deps := h.GetDeps(c)

	userId, err := uuid.NewV4()
	if err != nil {
		return err
	}
	if c.Stash().Get(shared.StashPathUserID).Exists() {
		userId, err = uuid.FromString(c.Stash().Get(shared.StashPathUserID).String())
		if err != nil {
			return fmt.Errorf("failed to parse stashed user_id into a uuid: %w", err)
		}
	}

	err = h.createUser(
		c,
		userId,
		c.Stash().Get(shared.StashPathEmail).String(),
		c.Stash().Get(shared.StashPathEmailVerified).Bool(),
		c.Stash().Get(shared.StashPathUsername).String(),
		c.Stash().Get(shared.StashPathWebauthnCredentials).Array(),
		c.Stash().Get(shared.StashPathNewPassword).String(),
		c.Stash().Get(shared.StashPathOTPSecret).String(),
		c.Stash().Get(shared.StashPathUsePasskeyForMFA).Bool(),
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	utils.NotifyUserChange(deps.HttpContext, deps.Tx, deps.Persister, events.UserCreate, userId)

	return nil
}

func (h CreateUser) createUser(c flowpilot.HookExecutionContext, id uuid.UUID, email string, emailVerified bool, username string, webauthnCredentials []gjson.Result, password, otpSecret string, usePasskeyForMFA bool) error {
	deps := h.GetDeps(c)

	now := time.Now().UTC()

	var auditLogDetails []auditlog.DetailOption

	err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Create(models.User{
		ID:               id,
		UsePasskeyForMFA: usePasskeyForMFA,
		CreatedAt:        now,
		UpdatedAt:        now,
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

	for _, webauthnCredential := range webauthnCredentials {
		var credentialModel models.WebauthnCredential
		err = json.Unmarshal([]byte(webauthnCredential.String()), &credentialModel)
		if err != nil {
			return fmt.Errorf("failed to unmarshal stashed webauthn_credential: %w", err)
		}

		err = deps.Persister.GetWebauthnCredentialPersisterWithConnection(deps.Tx).Create(credentialModel)
		if err != nil {
			return err
		}

		if credentialModel.MFAOnly {
			auditLogDetails = append(auditLogDetails, auditlog.Detail("security_key", credentialModel.ID))
		} else {
			auditLogDetails = append(auditLogDetails, auditlog.Detail("passkey", credentialModel.ID))
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

		auditLogDetails = append(auditLogDetails, auditlog.Detail("password", true))
	}

	if otpSecret != "" {
		otpSecretModel := models.NewOTPSecret(id, otpSecret)
		err = deps.Persister.GetOTPSecretPersisterWithConnection(deps.Tx).Create(*otpSecretModel)
		if err != nil {
			return err
		}

		auditLogDetails = append(auditLogDetails, auditlog.Detail("otp_secret", otpSecretModel.ID))
	}

	user, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(id)
	if err != nil {
		return err
	}

	if username != "" {
		usernameModel := models.NewUsername(user.ID, username)
		err = deps.Persister.GetUsernamePersisterWithConnection(deps.Tx).Create(*usernameModel)
		if err != nil {
			return err
		}
		auditLogDetails = append(auditLogDetails, auditlog.Detail("username", username))
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
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}
