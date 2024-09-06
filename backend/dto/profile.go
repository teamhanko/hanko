package dto

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type ProfileData struct {
	UserID       uuid.UUID                    `json:"user_id"`
	Passkeys     []WebauthnCredentialResponse `json:"passkeys,omitempty"`
	SecurityKeys []WebauthnCredentialResponse `json:"security_keys,omitempty"`
	Emails       []EmailResponse              `json:"emails,omitempty"`
	Username     *Username                    `json:"username,omitempty"`
	CreatedAt    time.Time                    `json:"created_at"`
	UpdatedAt    time.Time                    `json:"updated_at"`
}

func ProfileDataFromUserModel(user *models.User) *ProfileData {
	var webauthnCredentials, securityKeys []WebauthnCredentialResponse
	for _, webauthnCredentialModel := range user.WebauthnCredentials {
		webauthnCredential := FromWebauthnCredentialModel(&webauthnCredentialModel)
		if webauthnCredentialModel.MFAOnly {
			securityKeys = append(securityKeys, *webauthnCredential)
		} else {
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
		Emails:       emails,
		Username:     FromUsernameModel(user.Username),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}
