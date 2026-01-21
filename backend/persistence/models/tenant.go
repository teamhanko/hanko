package models

import (
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Tenant represents a tenant in a multi-tenant Hanko deployment
type Tenant struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Slug      string    `db:"slug" json:"slug"`
	Config    *string   `db:"config" json:"config,omitempty"`
	Enabled   bool      `db:"enabled" json:"enabled"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Tenants []Tenant

// NewTenant creates a new Tenant with the given ID, name, and slug
func NewTenant(id uuid.UUID, name, slug string) *Tenant {
	now := time.Now().UTC()
	return &Tenant{
		ID:        id,
		Name:      name,
		Slug:      slug,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewTenantWithGeneratedID creates a new Tenant with a generated UUID
func NewTenantWithGeneratedID(name, slug string) *Tenant {
	id, _ := uuid.NewV4()
	return NewTenant(id, name, slug)
}

// Validate gets run every time you call a "pop.Validate*" method
func (tenant *Tenant) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: tenant.ID},
		&validators.StringIsPresent{Name: "Name", Field: tenant.Name},
		&validators.StringIsPresent{Name: "Slug", Field: tenant.Slug},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: tenant.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: tenant.UpdatedAt},
	), nil
}
