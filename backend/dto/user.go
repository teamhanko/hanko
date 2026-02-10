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
	UserID     string       `json:"user_id"`
	Email      *EmailJWT    `json:"email,omitempty"`
	Username   string       `json:"username"`
	Metadata   *MetadataJWT `json:"metadata,omitempty"`
	Name       string       `json:"name"`
	FamilyName string       `json:"family_name"`
	GivenName  string       `json:"given_name"`
	Picture    string       `json:"picture"`
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

	if userModel.GivenName.Valid {
		userJWT.GivenName = userModel.GivenName.String
	}

	if userModel.FamilyName.Valid {
		userJWT.FamilyName = userModel.FamilyName.String
	}

	if userModel.Name.Valid {
		userJWT.Name = userModel.Name.String
	}

	if userModel.Picture.Valid {
		userJWT.Picture = userModel.Picture.String
	}

	return userJWT
}
