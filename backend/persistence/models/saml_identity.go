package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

type SamlIdentity struct {
	ID         uuid.UUID `json:"id" db:"id"`
	IdentityID uuid.UUID `json:"identity_id" db:"identity_id"`
	Domain     string    `json:"domain" db:"domain"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type SamlIdentities []SamlIdentity

func (i *SamlIdentity) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: i.ID},
		&validators.UUIDIsPresent{Name: "IdentityID", Field: i.IdentityID},
		&validators.StringIsPresent{Name: "Domain", Field: i.Domain},
	), nil
}
