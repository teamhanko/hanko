package dto

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type CreateUserResponse struct {
	ID      uuid.UUID `json:"id"` // deprecated
	UserID  uuid.UUID `json:"user_id"`
	EmailID uuid.UUID `json:"email_id"`
}

type GetUserResponse struct {
	ID                  uuid.UUID                   `json:"id"`
	Email               *string                     `json:"email,omitempty"`
	Username            *string                     `json:"username,omitempty"`
	WebauthnCredentials []models.WebauthnCredential `json:"webauthn_credentials"` // deprecated
	UpdatedAt           time.Time                   `json:"updated_at"`
	CreatedAt           time.Time                   `json:"created_at"`
}

type UserInfoResponse struct {
	ID                    uuid.UUID `json:"id"`
	EmailID               uuid.UUID `json:"email_id"`
	Verified              bool      `json:"verified"`
	HasWebauthnCredential bool      `json:"has_webauthn_credential"`
}

// UserJWT represents an abstracted user model for session management
type UserJWT struct {
	UserID   string
	Email    *EmailJWT
	Username string
}

func UserJWTFromUserModel(userModel *models.User) UserJWT {
	userJWT := UserJWT{
		UserID: userModel.ID.String(),
	}

	if primaryEmail := userModel.Emails.GetPrimary(); primaryEmail != nil {
		userJWT.Email = EmailJWTFromEmailModel(primaryEmail)
	}

	if userModel.Username != nil {
		userJWT.Username = userModel.Username.Username
	}

	return userJWT
}
