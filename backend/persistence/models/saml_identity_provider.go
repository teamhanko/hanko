package models

import (
	"github.com/gobuffalo/validate/v3/validators"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// SamlIdentityProvider is used by pop to map your saml_identity_providers database table to your go code.
type SamlIdentityProvider struct {
	ID                    uuid.UUID        `json:"id" db:"id"`
	Enabled               bool             `json:"enabled" db:"enabled"`
	Name                  string           `json:"name" db:"name"`
	Domain                string           `json:"domain" db:"domain"`
	MetadataUrl           string           `json:"metadata_url" db:"metadata_url"`
	SkipEmailVerification bool             `json:"skip_email_verification" db:"skip_email_verification"`
	AttributeMap          SamlAttributeMap `json:"attribute_map" has_one:"saml_attribute_map"`
	CreatedAt             time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time        `json:"updated_at" db:"updated_at"`
}

// SamlIdentityProviders is not required by pop and may be deleted
type SamlIdentityProviders []SamlIdentityProvider

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (s *SamlIdentityProvider) Validate(_ *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: s.ID},
		&validators.StringIsPresent{Name: "Name", Field: s.Name},
		&validators.StringIsPresent{Name: "Domain", Field: s.Domain},
		&validators.URLIsPresent{Name: "Metadata", Field: s.MetadataUrl},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: s.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: s.CreatedAt},
	), nil
}
