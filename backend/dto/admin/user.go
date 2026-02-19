package admin

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
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
	OTP                 *OTPDto                          `json:"otp,omitempty"`
	IPAddress           *string                          `json:"ip_address,omitempty"`
	UserAgent           *string                          `json:"user_agent,omitempty"`
	Metadata            *Metadata                        `json:"metadata,omitempty"`
	GivenName           string                           `json:"given_name,omitempty"`
	FamilyName          string                           `json:"family_name,omitempty"`
	Name                string                           `json:"name,omitempty"`
	Picture             string                           `json:"picture,omitempty"`
}

func (u *User) SetIPAddress(ip string) {
	u.IPAddress = &ip
}

func (u *User) SetUserAgent(agent string) {
	u.UserAgent = &agent
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

	var otp *OTPDto = nil
	if model.OTPSecret != nil {
		otp = &OTPDto{
			ID:        model.OTPSecret.ID,
			CreatedAt: model.OTPSecret.CreatedAt,
		}
	}

	var metadata *Metadata
	if model.Metadata != nil {
		metadata = NewMetadata(model.Metadata)
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
		OTP:                 otp,
		Metadata:            metadata,
		GivenName:           model.GivenName.String,
		FamilyName:          model.FamilyName.String,
		Name:                model.Name.String,
		Picture:             model.Picture.String,
	}
}

type CreateUser struct {
	ID        uuid.UUID     `json:"id"`
	Emails    []CreateEmail `json:"emails" validate:"unique=Address,dive"`
	Username  *string       `json:"username"`
	CreatedAt time.Time     `json:"created_at"`
}
