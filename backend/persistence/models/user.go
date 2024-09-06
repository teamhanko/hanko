package models

import (
	"encoding/base64"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"golang.org/x/exp/slices"
	"time"
)

// User is used by pop to map your users database table to your go code.
type User struct {
	ID                  uuid.UUID           `db:"id" json:"id"`
	WebauthnCredentials WebauthnCredentials `has_many:"webauthn_credentials" json:"webauthn_credentials,omitempty"`
	Emails              Emails              `has_many:"emails" json:"-"`
	CreatedAt           time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time           `db:"updated_at" json:"updated_at"`
	Username            *Username           `has_one:"username" json:"username,omitempty"`
	OTPSecret           *OTPSecret          `has_one:"otp_secret" json:"-"`
	PasswordCredential  *PasswordCredential `has_one:"password_credentials" json:"-"`
}

func (user *User) DeleteWebauthnCredential(credentialId string) {
	for i := range user.WebauthnCredentials {
		if user.WebauthnCredentials[i].ID == credentialId {
			user.WebauthnCredentials = slices.Delete(user.WebauthnCredentials, i, i+1)
			return
		}
	}
}

func (user *User) GetIdentities() Identities {
	var identities Identities
	for _, email := range user.Emails {
		identities = append(identities, email.Identities...)
	}
	return identities
}

func NewUser() User {
	id, _ := uuid.NewV4()
	return User{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (user *User) GetUsername() *string {
	if user.Username != nil {
		return &user.Username.Username
	}
	return nil
}

func (user *User) SetUsername(username *Username) {
	user.Username = username
}

func (user *User) DeleteUsername() {
	user.Username = nil
}

func (user *User) SetPrimaryEmail(primary *PrimaryEmail) {
	for i := range user.Emails {
		if user.Emails[i].ID.String() == primary.EmailID.String() {
			user.Emails[i].PrimaryEmail = primary
		} else {
			user.Emails[i].PrimaryEmail = nil
		}
	}
}

func (user *User) UpdateEmail(email Email) {
	for i := range user.Emails {
		if user.Emails[i].ID.String() == email.ID.String() {
			user.Emails[i] = email
			return
		}
	}
}

func (user *User) DeleteEmail(email Email) {
	for i := range user.Emails {
		if user.Emails[i].ID.String() == email.ID.String() {
			user.Emails = slices.Delete(user.Emails, i, i+1)
			return
		}
	}
}

func (user *User) DeleteOTPSecret() {
	user.OTPSecret = nil
}

func (user *User) GetEmailById(emailId uuid.UUID) *Email {
	return user.Emails.GetEmailById(emailId)
}

func (user *User) GetEmailByAddress(address string) *Email {
	return user.Emails.GetEmailByAddress(address)
}

func (user *User) GetWebauthnCredentialById(credentialId string) *WebauthnCredential {
	for i := range user.WebauthnCredentials {
		if user.WebauthnCredentials[i].ID == credentialId {
			return &user.WebauthnCredentials[i]
		}
	}
	return nil
}

func (user *User) GetPasskeys() WebauthnCredentials {
	credentials := make(WebauthnCredentials, 0)
	for _, credential := range user.WebauthnCredentials {
		if credential.MFAOnly == false {
			credentials = append(credentials, credential)
		}
	}
	return credentials
}

func (user *User) GetSecurityKeys() WebauthnCredentials {
	credentials := make(WebauthnCredentials, 0)
	for _, credential := range user.WebauthnCredentials {
		if credential.MFAOnly == true {
			credentials = append(credentials, credential)
		}
	}
	return credentials
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
