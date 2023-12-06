package models

import (
	"encoding/base64"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

// User is used by pop to map your users database table to your go code.
type User struct {
	ID                  uuid.UUID            `db:"id" json:"id"`
	WebauthnCredentials []WebauthnCredential `has_many:"webauthn_credentials" json:"webauthn_credentials,omitempty"`
	Emails              Emails               `has_many:"emails" json:"emails"`
	Username            string               `db:"username" json:"username,omitempty"`
	CreatedAt           time.Time            `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time            `db:"updated_at" json:"updated_at"`
}

func NewUser() User {
	id, _ := uuid.NewV4()
	return User{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (user *User) GetEmailById(emailId uuid.UUID) *Email {
	for _, email := range user.Emails {
		if email.ID.String() == emailId.String() {
			return &email
		}
	}
	return nil
}

func (user *User) GetWebauthnCredentialById(credentialId string) *WebauthnCredential {
	for i := range user.WebauthnCredentials {
		if user.WebauthnCredentials[i].ID == credentialId {
			return &user.WebauthnCredentials[i]
		}
	}
	return nil
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (user *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: user.ID},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: user.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: user.CreatedAt},
	), nil
}

func (user *User) WebAuthnID() []byte {
	return user.ID.Bytes()
}

func (user *User) WebAuthnName() string {
	email := user.Emails.GetPrimary()
	if email != nil {
		return email.Address
	}
	return "username" // TODO
}

func (user *User) WebAuthnDisplayName() string {
	email := user.Emails.GetPrimary()
	if email != nil {
		return email.Address
	}
	return "username" // TODO
}

func (user *User) WebAuthnIcon() string {
	return ""
}

func (user *User) WebAuthnCredentials() []webauthn.Credential {
	var credentials []webauthn.Credential

	for _, credential := range user.WebauthnCredentials {
		credentialID, _ := base64.RawURLEncoding.DecodeString(credential.ID)
		publicKey, _ := base64.RawURLEncoding.DecodeString(credential.PublicKey)

		transport := make([]protocol.AuthenticatorTransport, len(credential.Transports))

		for i, t := range credential.Transports {
			transport[i] = protocol.AuthenticatorTransport(t.Name)
		}

		c := webauthn.Credential{
			ID:              credentialID,
			PublicKey:       publicKey,
			AttestationType: credential.AttestationType,
			Authenticator: webauthn.Authenticator{
				AAGUID:    credential.AAGUID.Bytes(),
				SignCount: uint32(credential.SignCount),
			},
			Transport: transport,
		}

		credentials = append(credentials, c)
	}

	return credentials
}
