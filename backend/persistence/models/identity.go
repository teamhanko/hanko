package models

import (
	"errors"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/pop/v6/slices"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Identity is used by pop to map your identities database table to your go code.
type Identity struct {
	ID             uuid.UUID     `json:"id" db:"id"`
	TenantID       *uuid.UUID    `json:"tenant_id,omitempty" db:"tenant_id"`
	ProviderUserID string        `json:"provider_user_id" db:"provider_user_id"`
	ProviderID     string        `json:"provider_id" db:"provider_id"`
	Data           slices.Map    `json:"data" db:"data"`
	EmailID        *uuid.UUID    `json:"email_id" db:"email_id"`
	UserID         *uuid.UUID    `json:"user_id" db:"user_id"`
	Email          *Email        `json:"email,omitempty" belongs_to:"email"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
	SamlIdentity   *SamlIdentity `json:"saml_identity" has_one:"saml_identity"`
}

type Identities []Identity

func (identities Identities) GetIdentity(providerID string, providerUserID string) *Identity {
	for _, identity := range identities {
		if identity.ProviderID == providerID && identity.ProviderUserID == providerUserID {
			return &identity
		}
	}

	return nil
}

func NewIdentity(providerID string, identityData map[string]interface{}, emailID *uuid.UUID, userID *uuid.UUID) (*Identity, error) {
	providerUserID, ok := identityData["sub"]
	if !ok {
		return nil, errors.New("missing provider user id")
	}
	now := time.Now().UTC()

	id, _ := uuid.NewV4()
	identity := &Identity{
		ID:             id,
		Data:           identityData,
		ProviderUserID: providerUserID.(string),
		ProviderID:     providerID,
		EmailID:        emailID,
		UserID:         userID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return identity, nil
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (i *Identity) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: i.ID},
		&validators.StringIsPresent{Name: "ProviderUserID", Field: i.ProviderUserID},
		&validators.StringIsPresent{Name: "ProviderID", Field: i.ProviderID},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: i.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: i.UpdatedAt},
	), nil
}
