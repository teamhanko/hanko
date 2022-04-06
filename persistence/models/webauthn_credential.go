package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

// WebauthnCredential is used by pop to map your webauthn_credentials database table to your go code.
type WebauthnCredential struct {
	ID              string    `db:"id"`
	UserId          uuid.UUID `db:"user_id"`
	PublicKey       string    `db:"public_key"`
	AttestationType string    `db:"attestation_type"`
	AAGUID          uuid.UUID `db:"aaguid"`
	SignCount       int       `db:"sign_count"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

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
