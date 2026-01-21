package admin

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

// Tenant represents a tenant in the admin API responses
type Tenant struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Config    *string   `json:"config,omitempty"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FromTenantModel converts the DB model to a DTO object
func FromTenantModel(model models.Tenant) Tenant {
	return Tenant{
		ID:        model.ID,
		Name:      model.Name,
		Slug:      model.Slug,
		Config:    model.Config,
		Enabled:   model.Enabled,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

// CreateTenant is the request body for creating a new tenant
type CreateTenant struct {
	ID      *uuid.UUID `json:"id,omitempty"`
	Name    string     `json:"name" validate:"required,min=1,max=255"`
	Slug    string     `json:"slug" validate:"required,min=1,max=255"`
	Config  *string    `json:"config,omitempty"`
	Enabled *bool      `json:"enabled,omitempty"`
}

// UpdateTenant is the request body for updating an existing tenant
type UpdateTenant struct {
	Name    *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Slug    *string `json:"slug,omitempty" validate:"omitempty,min=1,max=255"`
	Config  *string `json:"config,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}
