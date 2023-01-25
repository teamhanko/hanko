package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// WebauthnCredentialTransport is used by pop to map your webauthn_credential_transport table to your go code.
type WebauthnCredentialTransport struct {
	ID                   uuid.UUID           `db:"id"`
	Name                 string              `db:"name"`
	WebauthnCredentialID string              `db:"webauthn_credential_id"`
	WebauthnCredential   *WebauthnCredential `belongs_to:"webauthn_credential"`
}

type Transports []WebauthnCredentialTransport

func (transports Transports) GetNames() []string {
	names := make([]string, len(transports))
	for i, t := range transports {
		names[i] = t.Name
	}
	return names
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (transport *WebauthnCredentialTransport) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: transport.ID},
		&validators.StringIsPresent{Name: "WebauthnCredentialID", Field: transport.WebauthnCredentialID},
		&validators.StringIsPresent{Name: "Name", Field: transport.Name},
	), nil
}
