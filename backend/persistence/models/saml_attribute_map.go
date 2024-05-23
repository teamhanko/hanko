package models

import (
	"github.com/gobuffalo/validate/v3/validators"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// SamlAttributeMap is used by pop to map your saml_attribute_maps database table to your go code.
type SamlAttributeMap struct {
	ID                 uuid.UUID             `json:"id" db:"id"`
	IdentityProviderID uuid.UUID             `json:"-" db:"saml_identity_provider_id"`
	IdentityProvider   *SamlIdentityProvider `json:"-" belongs_to:"saml_identity_provider"`
	Name               string                `json:"name,omitempty" db:"name"`
	FamilyName         string                `json:"family_name,omitempty" db:"family_name"`
	GivenName          string                `json:"given_name,omitempty" db:"given_name"`
	MiddleName         string                `json:"middle_name,omitempty" db:"middle_name"`
	NickName           string                `json:"nickname,omitempty" db:"nickname"`
	PreferredUsername  string                `json:"preferred_username,omitempty" db:"preferred_username"`
	Profile            string                `json:"profile,omitempty" db:"profile"`
	Picture            string                `json:"picture,omitempty" db:"picture"`
	Website            string                `json:"website,omitempty" db:"website"`
	Gender             string                `json:"gender,omitempty" db:"gender"`
	Birthdate          string                `json:"birthdate,omitempty" db:"birthdate"`
	ZoneInfo           string                `json:"zone_info,omitempty" db:"zone_info"`
	Locale             string                `json:"locale,omitempty" db:"locale"`
	SamlUpdatedAt      string                `json:"saml_updated_at,omitempty" db:"saml_updated_at"`
	Email              string                `json:"email,omitempty" db:"email"`
	EmailVerified      string                `json:"email_verified,omitempty" db:"email_verified"`
	Phone              string                `json:"phone,omitempty" db:"phone"`
	PhoneVerified      string                `json:"phone_verified,omitempty" db:"phone_verified"`
	CreatedAt          time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time             `json:"updated_at" db:"updated_at"`
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (s *SamlAttributeMap) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: s.ID},
	), nil
}
