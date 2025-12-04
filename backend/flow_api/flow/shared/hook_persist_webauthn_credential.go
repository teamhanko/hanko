package shared

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/v2/audit_log"
	"github.com/teamhanko/hanko/backend/v2/flow_api/services"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
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

		var userModel *models.User
		if userValue, ok := c.Get("session_user").(*models.User); ok {
			userModel = userValue
			userModel.WebauthnCredentials = append(userModel.WebauthnCredentials, credentialModel)
		}

		var isPasskey bool = false

		if credentialModel.MFAOnly {
			auditLogType = models.AuditLogSecurityKeyCreated
			auditLogDetails = append(auditLogDetails, auditlog.Detail("security_key", credentialModel.ID))
		} else {
			isPasskey = true
			auditLogDetails = append(auditLogDetails, auditlog.Detail("passkey", credentialModel.ID))
		}

		if userModel != nil {
			emailAddress := userModel.Emails.GetPrimary().Address

			if !isPasskey {
				var hasOtherMfa bool = false

				for _, credential := range userModel.WebauthnCredentials {
					if credential.ID != credentialModel.ID && credential.MFAOnly {
						// User has another MFA-only credential
						hasOtherMfa = true
						break
					}
				}

				// Send MFA enabled notification if this is the first MFA method
				if !hasOtherMfa && userModel.OTPSecret == nil && deps.Cfg.SecurityNotifications.Notifications.MFAEnabled.Enabled {
					deps.SecurityNotificationService.SendNotification(deps.Tx, services.SendSecurityNotificationParams{
						EmailAddress: emailAddress,
						Template:     "mfa_enabled",
						HttpContext:  deps.HttpContext,
						UserContext:  *userModel,
					})
				}
			}

			if isPasskey && deps.Cfg.SecurityNotifications.Notifications.PasskeyCreate.Enabled {
				deps.SecurityNotificationService.SendNotification(deps.Tx, services.SendSecurityNotificationParams{
					EmailAddress: emailAddress,
					Template:     "passkey_create",
					HttpContext:  deps.HttpContext,
					UserContext:  *userModel,
				})
			}
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
