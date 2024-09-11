package models

import (
	"encoding/base64"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

// WebauthnCredential is used by pop to map your webauthn_credentials database table to your go code.
type WebauthnCredential struct {
	ID              string     `db:"id" json:"id"`
	Name            *string    `db:"name" json:"-"`
	UserId          uuid.UUID  `db:"user_id" json:"-"`
	PublicKey       string     `db:"public_key" json:"-"`
	AttestationType string     `db:"attestation_type" json:"-"`
	AAGUID          uuid.UUID  `db:"aaguid" json:"-"`
	SignCount       int        `db:"sign_count" json:"-"`
	LastUsedAt      *time.Time `db:"last_used_at" json:"-"`
	CreatedAt       time.Time  `db:"created_at" json:"-"`
	UpdatedAt       time.Time  `db:"updated_at" json:"-"`
	Transports      Transports `has_many:"webauthn_credential_transports" json:"-"`
	BackupEligible  bool       `db:"backup_eligible" json:"-"`
	BackupState     bool       `db:"backup_state" json:"-"`
	MFAOnly         bool       `db:"mfa_only" json:"-"`
}

type WebauthnCredentials []WebauthnCredential

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (credential *WebauthnCredential) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Name: "ID", Field: credential.ID},
		&validators.UUIDIsPresent{Name: "UserId", Field: credential.UserId},
		&validators.StringIsPresent{Name: "PublicKey", Field: credential.PublicKey},
		&validators.IntIsGreaterThan{Name: "SignCount", Field: credential.SignCount, Compared: -1},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: credential.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: credential.UpdatedAt},
	), nil
}

func (credential *WebauthnCredential) GetWebauthnTransports() []protocol.AuthenticatorTransport {
	transports := make([]protocol.AuthenticatorTransport, len(credential.Transports))
	for i, transport := range credential.Transports {
		transports[i] = protocol.AuthenticatorTransport(transport.Name)
	}
	return transports
}

func (credential *WebauthnCredential) GetWebauthnDescriptor() (*protocol.CredentialDescriptor, error) {
	id, err := base64.RawURLEncoding.DecodeString(credential.ID)
	if err != nil {
		fmt.Println("failed to decode the credential id", err)
		return nil, err
	}

	return &protocol.CredentialDescriptor{
		Type:            protocol.PublicKeyCredentialType,
		CredentialID:    id,
		Transport:       credential.GetWebauthnTransports(),
		AttestationType: credential.AttestationType,
	}, nil
}

func (credentials WebauthnCredentials) GetWebauthnDescriptors() ([]protocol.CredentialDescriptor, error) {
	descriptors := make([]protocol.CredentialDescriptor, len(credentials))
	for i, credential := range credentials {
		descriptor, err := credential.GetWebauthnDescriptor()
		if err != nil {
			return nil, err
		}
		descriptors[i] = *descriptor
	}
	return descriptors, nil
}
