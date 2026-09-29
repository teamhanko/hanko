package admin

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type Organization struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FromOrganizationModel converts the DB model to a DTO object
func FromOrganizationModel(model models.Organization) Organization {
	return Organization{
		ID:        model.ID,
		Name:      model.Name,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

type CreateOrganization struct {
	Name string `json:"name" validate:"required"`
}

type PatchOrganizationRequest struct {
	Name dto.OptionalString `json:"name"`
}
