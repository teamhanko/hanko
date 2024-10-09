package dto

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type MFAConfig struct {
	AuthAppSetUp        bool `json:"auth_app_set_up"`
	TOTPEnabled         bool `json:"totp_enabled"`
	SecurityKeysEnabled bool `json:"security_keys_enabled"`
}

type ProfileData struct {
	UserID       uuid.UUID                    `json:"user_id"`
	Passkeys     []WebauthnCredentialResponse `json:"passkeys,omitempty"`
	SecurityKeys []WebauthnCredentialResponse `json:"security_keys,omitempty"`
	MFAConfig    MFAConfig                    `json:"mfa_config"`
	Emails       []EmailResponse              `json:"emails,omitempty"`
	Username     *Username                    `json:"username,omitempty"`
	CreatedAt    time.Time                    `json:"created_at"`
	UpdatedAt    time.Time                    `json:"updated_at"`
}

func ProfileDataFromUserModel(user *models.User, cfg *config.Config) *ProfileData {
	var webauthnCredentials, securityKeys []WebauthnCredentialResponse
	for _, webauthnCredentialModel := range user.WebauthnCredentials {
		webauthnCredential := FromWebauthnCredentialModel(&webauthnCredentialModel)
		if cfg.MFA.SecurityKeys.Enabled && webauthnCredentialModel.MFAOnly {
			securityKeys = append(securityKeys, *webauthnCredential)
		} else if cfg.Passkey.Enabled {
			webauthnCredentials = append(webauthnCredentials, *webauthnCredential)
		}
	}

	var emails []EmailResponse
	for _, emailModel := range user.Emails {
		email := FromEmailModel(&emailModel)
		emails = append(emails, *email)
	}

	return &ProfileData{
		UserID:       user.ID,
		Passkeys:     webauthnCredentials,
		SecurityKeys: securityKeys,
		MFAConfig: MFAConfig{
			AuthAppSetUp:        user.OTPSecret != nil,
			TOTPEnabled:         cfg.MFA.Enabled && cfg.MFA.TOTP.Enabled,
			SecurityKeysEnabled: cfg.MFA.Enabled && cfg.MFA.SecurityKeys.Enabled,
		},
		Emails:    emails,
		Username:  FromUsernameModel(user.Username),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
