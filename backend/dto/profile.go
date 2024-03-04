package dto

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type ProfileData struct {
	UserID              uuid.UUID                    `json:"user_id"`
	WebauthnCredentials []WebauthnCredentialResponse `json:"passkeys,omitempty"`
	Emails              []EmailResponse              `json:"emails,omitempty"`
	Username            string                       `json:"username,omitempty"`
	CreatedAt           time.Time                    `json:"created_at"`
	UpdatedAt           time.Time                    `json:"updated_at"`
}

func ProfileDataFromUserModel(user *models.User) *ProfileData {
	var webauthnCredentials []WebauthnCredentialResponse
	for _, webauthnCredentialModel := range user.WebauthnCredentials {
		webauthnCredential := FromWebauthnCredentialModel(&webauthnCredentialModel)
		webauthnCredentials = append(webauthnCredentials, *webauthnCredential)
	}

	var emails []EmailResponse
	for _, emailModel := range user.Emails {
		email := FromEmailModel(&emailModel)
		emails = append(emails, *email)
	}

	return &ProfileData{
		UserID:              user.ID,
		WebauthnCredentials: webauthnCredentials,
		Emails:              emails,
		Username:            user.Username,
		CreatedAt:           user.CreatedAt,
		UpdatedAt:           user.UpdatedAt,
	}
}
