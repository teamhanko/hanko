package models

import (
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Role is used by pop to map your roles database table to your go code.
type Role struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TenantID  uuid.UUID `json:"-" db:"tenant_id"`
	Slug      string    `json:"slug" db:"slug"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Roles []Role

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (r *Role) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: r.ID},
		&validators.UUIDIsPresent{Name: "TenantID", Field: r.TenantID},
		&validators.StringIsPresent{Name: "Slug", Field: r.Slug},
		&validators.StringIsPresent{Name: "Name", Field: r.Name},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: r.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: r.UpdatedAt},
	), nil
}
