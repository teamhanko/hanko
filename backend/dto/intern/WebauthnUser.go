package intern

import (
	"errors"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewWebauthnUser(user models.User, credentials []models.WebauthnCredential) (*WebauthnUser, error) {
	email := user.Emails.GetPrimary()
	if email == nil {
		return nil, errors.New("primary email unavailable")
	}

	return &WebauthnUser{
		UserId:              user.ID,
		Email:               email.Address,
		WebauthnCredentials: credentials,
	}, nil
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
		cred := credential
		c := WebauthnCredentialFromModel(&cred)
		credentials = append(credentials, *c)
	}

	return credentials
}
