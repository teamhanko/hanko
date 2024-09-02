package admin

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type Identity struct {
	ID           uuid.UUID `json:"id"`
	ProviderID   string    `json:"provider_id"`
	ProviderName string    `json:"provider_name"`
	EmailID      uuid.UUID `json:"email_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func FromIdentityModel(model models.Identity) Identity {
	return Identity{
		ID:           model.ID,
		ProviderID:   model.ProviderID,
		ProviderName: model.ProviderName,
		EmailID:      model.EmailID,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}
