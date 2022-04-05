package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

// WebauthnSessionDataAllowedCredential is used by pop to map your webauthn_session_data_allowed_credential database table to your go code.
type WebauthnSessionDataAllowedCredential struct {
	ID                    uuid.UUID            `db:"id"`
	CredentialId          string               `db:"credential_id"`
	WebauthnSessionDataID uuid.UUID            `db:"session_data_id"`
	CreatedAt             time.Time            `db:"created_at"`
	UpdatedAt             time.Time            `db:"updated_at"`
	WebauthnSessionData   *WebauthnSessionData `belongs_to:"webauthn_session_data"`
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (credential *WebauthnSessionDataAllowedCredential) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: credential.ID},
		&validators.StringLengthInRange{Name: "CredentialId", Field: credential.CredentialId, Min: 1, Max: 0},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: credential.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: credential.CreatedAt},
	), nil
}
