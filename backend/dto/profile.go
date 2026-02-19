package dto

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
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
	Metadata     *Metadata                    `json:"metadata,omitempty"`
	Identities   Identities                   `json:"identities,omitempty"`
	Name         string                       `json:"name,omitempty"`
	GivenName    string                       `json:"given_name,omitempty"`
	FamilyName   string                       `json:"family_name,omitempty"`
	Picture      string                       `json:"picture,omitempty"`
}

func ProfileDataFromUserModel(user *models.User, cfg *config.Config) *ProfileData {
	var webauthnCredentials, securityKeys []WebauthnCredentialResponse
	for _, webauthnCredentialModel := range user.WebauthnCredentials {
		webauthnCredential := FromWebauthnCredentialModel(&webauthnCredentialModel)
		if cfg.MFA.SecurityKeys.Enabled && webauthnCredentialModel.MFAOnly {
			securityKeys = append(securityKeys, *webauthnCredential)
		} else if cfg.Passkey.Enabled && !webauthnCredentialModel.MFAOnly {
			webauthnCredentials = append(webauthnCredentials, *webauthnCredential)
		}
	}

	var emails []EmailResponse
	for _, emailModel := range user.Emails {
		email := FromEmailModel(&emailModel, cfg)
		emails = append(emails, *email)
	}

	var metadata *Metadata
	if user.Metadata != nil {
		metadata = NewMetadata(user.Metadata)
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
		Emails:     emails,
		Username:   FromUsernameModel(user.Username),
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
		Metadata:   metadata,
		Identities: FromIdentitiesModel(user.Identities, cfg),
		Name:       user.Name.String,
		GivenName:  user.GivenName.String,
		FamilyName: user.FamilyName.String,
		Picture:    user.Picture.String,
	}
}
