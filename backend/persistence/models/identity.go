package models

import (
	"errors"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/pop/v6/slices"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

// Identity is used by pop to map your identities database table to your go code.
type Identity struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	ProviderID   string     `json:"provider_id" db:"provider_id"`
	ProviderName string     `json:"provider_name" db:"provider_name"`
	Data         slices.Map `json:"data" db:"data"`
	EmailID      uuid.UUID  `json:"email_id" db:"email_id"`
	Email        *Email     `json:"email" belongs_to:"email"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type Identities []Identity

func NewIdentity(provider string, identityData map[string]interface{}, emailID uuid.UUID) (*Identity, error) {
	providerID, ok := identityData["sub"]
	if !ok {
		return nil, errors.New("missing provider id")
	}
	now := time.Now().UTC()

	id, _ := uuid.NewV4()
	identity := &Identity{
		ID:           id,
		Data:         identityData,
		ProviderID:   providerID.(string),
		ProviderName: provider,
		EmailID:      emailID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return identity, nil
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (i *Identity) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: i.ID},
		&validators.StringIsPresent{Name: "ProviderID", Field: i.ProviderID},
		&validators.StringIsPresent{Name: "ProviderName", Field: i.ProviderName},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: i.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: i.UpdatedAt},
	), nil
}
