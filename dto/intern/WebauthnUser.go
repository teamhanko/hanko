package intern

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/persistence/models"
)

func NewWebauthnUser(user models.User, credentials []models.WebauthnCredential) *WebauthnUser {
	return &WebauthnUser{
		UserId:              user.ID,
		Email:               user.Email,
		WebauthnCredentials: credentials,
	}
}

type WebauthnUser struct {
	UserId              uuid.UUID
	Email               string
	WebauthnCredentials []models.WebauthnCredential
}

func (u *WebauthnUser) WebAuthnID() []byte {
	return u.UserId.Bytes()
}

func (u *WebauthnUser) WebAuthnName() string {
	return u.Email
}

func (u *WebauthnUser) WebAuthnDisplayName() string {
	return u.Email
}

func (u *WebauthnUser) WebAuthnIcon() string {
	return ""
}

func (u *WebauthnUser) WebAuthnCredentials() []webauthn.Credential {
	var credentials []webauthn.Credential
	for _, credential := range u.WebauthnCredentials {
		c := WebauthnCredentialFromModel(&credential)
		credentials = append(credentials, *c)
	}

	return credentials
}
