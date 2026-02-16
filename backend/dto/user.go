package dto

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
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
	Metadata            *Metadata                   `json:"metadata,omitempty"`
	ProfileData         `json:",inline"`
}

type UserInfoResponse struct {
	ID                    uuid.UUID `json:"id"`
	EmailID               uuid.UUID `json:"email_id"`
	Verified              bool      `json:"verified"`
	HasWebauthnCredential bool      `json:"has_webauthn_credential"`
}

// UserJWT represents an abstracted user model for session management
type UserJWT struct {
	UserID   string       `json:"user_id"`
	TenantID *string      `json:"tenant_id,omitempty"`
	Email    *EmailJWT    `json:"email,omitempty"`
	Username string       `json:"username"`
	Metadata *MetadataJWT `json:"metadata,omitempty"`
}

func (u *UserJWT) String() string {
	if u == nil {
		return ""
	}

	jsonBytes, _ := json.Marshal(u)
	return string(jsonBytes)
}

func UserJWTFromUserModel(userModel *models.User) UserJWT {
	userJWT := UserJWT{
		UserID: userModel.ID.String(),
	}

	if userModel.TenantID != nil {
		tenantID := userModel.TenantID.String()
		userJWT.TenantID = &tenantID
	}

	if primaryEmail := userModel.Emails.GetPrimary(); primaryEmail != nil {
		userJWT.Email = EmailJWTFromEmailModel(primaryEmail)
	}

	if userModel.Username != nil {
		userJWT.Username = userModel.Username.Username
	}

	if userModel.Metadata != nil {
		metadataJWT := MetadataJWTFromUserModel(userModel.Metadata)
		if metadataJWT != nil {
			userJWT.Metadata = metadataJWT
		}
	}

	return userJWT
}
