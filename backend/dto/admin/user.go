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
	Username            *Username                        `json:"username,omitempty"`
	CreatedAt           time.Time                        `json:"created_at"`
	UpdatedAt           time.Time                        `json:"updated_at"`
	Password            *PasswordCredential              `json:"password,omitempty"`
	Identities          []Identity                       `json:"identities,omitempty"`
}

// FromUserModel Converts the DB model to a DTO object
func FromUserModel(model models.User) User {
	credentials := make([]dto.WebauthnCredentialResponse, len(model.WebauthnCredentials))
	for i := range model.WebauthnCredentials {
		credentials[i] = *dto.FromWebauthnCredentialModel(&model.WebauthnCredentials[i])
	}
	emails := make([]Email, len(model.Emails))
	var identities = make([]Identity, 0)
	for i := range model.Emails {
		emails[i] = *FromEmailModel(&model.Emails[i])
		for j := range model.Emails[i].Identities {
			identities = append(identities, FromIdentityModel(model.Emails[i].Identities[j]))
		}
	}
	var username *Username = nil
	if model.Username != nil {
		username = FromUsernameModel(model.Username)
	}

	var passwordCredential *PasswordCredential = nil
	if model.PasswordCredential != nil {
		passwordCredential = &PasswordCredential{
			ID:        model.PasswordCredential.ID,
			CreatedAt: model.PasswordCredential.CreatedAt,
			UpdatedAt: model.PasswordCredential.UpdatedAt,
		}
	}

	return User{
		ID:                  model.ID,
		WebauthnCredentials: credentials,
		Emails:              emails,
		Username:            username,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
		Password:            passwordCredential,
		Identities:          identities,
	}
}

type CreateUser struct {
	ID        uuid.UUID     `json:"id"`
	Emails    []CreateEmail `json:"emails" validate:"unique=Address,dive"`
	Username  *string       `json:"username"`
	CreatedAt time.Time     `json:"created_at"`
}
