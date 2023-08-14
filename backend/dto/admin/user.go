package admin

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type User struct {
	ID                  uuid.UUID                        `json:"id"`
	WebauthnCredentials []dto.WebauthnCredentialResponse `json:"webauthn_credentials,omitempty"`
	Emails              []Email                          `json:"emails,omitempty"`
	CreatedAt           time.Time                        `json:"created_at"`
	UpdatedAt           time.Time                        `json:"updated_at"`
}

// FromUserModel Converts the DB model to a DTO object
func FromUserModel(model models.User) User {
	credentials := make([]dto.WebauthnCredentialResponse, len(model.WebauthnCredentials))
	for i := range model.WebauthnCredentials {
		credentials[i] = *dto.FromWebauthnCredentialModel(&model.WebauthnCredentials[i])
	}
	emails := make([]Email, len(model.Emails))
	for i := range model.Emails {
		emails[i] = *FromEmailModel(&model.Emails[i])
	}
	return User{
		ID:                  model.ID,
		WebauthnCredentials: credentials,
		Emails:              emails,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
}

type CreateUser struct {
	ID        uuid.UUID     `json:"id"`
	Emails    []CreateEmail `json:"emails" validate:"required,gte=1,unique=Address,dive"`
	CreatedAt time.Time     `json:"created_at"`
}
