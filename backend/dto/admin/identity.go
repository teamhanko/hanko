package admin

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type Identity struct {
	ID           uuid.UUID  `json:"id"`
	ProviderID   string     `json:"provider_id"`
	ProviderName string     `json:"provider_name"`
	EmailID      *uuid.UUID `json:"email_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func FromIdentityModel(model models.Identity) Identity {
	return Identity{
		ID:           model.ID,
		ProviderID:   model.ProviderUserID,
		ProviderName: model.ProviderID,
		EmailID:      model.EmailID,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}
