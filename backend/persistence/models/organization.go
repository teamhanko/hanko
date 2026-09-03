package models

import (
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Organization is used by pop to map your organizations database table to your go code.
type Organization struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TenantID  uuid.UUID `json:"-" db:"tenant_id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Organizations []Organization

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (o *Organization) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: o.ID},
		&validators.UUIDIsPresent{Name: "TenantID", Field: o.TenantID},
		&validators.StringIsPresent{Name: "Name", Field: o.Name},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: o.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: o.UpdatedAt},
	), nil
}
