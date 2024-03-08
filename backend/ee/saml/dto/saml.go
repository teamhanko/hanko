package dto

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type SamlRequest struct {
	Domain string `query:"domain" validate:"required,fqdn"`
}

type SamlMetadataRequest struct {
	SamlRequest
	CertOnly bool `query:"cert_only" validate:"boolean"`
}

type SamlAuthRequest struct {
	SamlRequest
	RedirectTo string `query:"redirect_to" validate:"required,url"`
}

type SamlCreateProviderAttributeMapRequest struct {
	Name              string `json:"name" validate:"omitempty"`
	FamilyName        string `json:"family_name" validate:"omitempty"`
	GivenName         string `json:"given_name" validate:"omitempty"`
	MiddleName        string `json:"middle_name" validate:"omitempty"`
	NickName          string `json:"nickname" validate:"omitempty"`
	PreferredUsername string `json:"preferred_username" validate:"omitempty"`
	Profile           string `json:"profile" validate:"omitempty"`
	Picture           string `json:"picture" validate:"omitempty"`
	Website           string `json:"website" validate:"omitempty"`
	Gender            string `json:"gender" validate:"omitempty"`
	Birthdate         string `json:"birthdate" validate:"omitempty"`
	ZoneInfo          string `json:"zone_info" validate:"omitempty"`
	Locale            string `json:"locale" validate:"omitempty"`
	UpdatedAt         string `json:"updated_at" validate:"omitempty"`
	Email             string `json:"email" validate:"omitempty"`
	EmailVerified     string `json:"email_verified" validate:"omitempty"`
	Phone             string `json:"phone" validate:"omitempty"`
	PhoneVerified     string `json:"phone_verified" validate:"omitempty"`
}

func (sam *SamlCreateProviderAttributeMapRequest) ToModel(model *models.SamlAttributeMap) *models.SamlAttributeMap {
	model.Name = sam.Name
	model.FamilyName = sam.FamilyName
	model.GivenName = sam.GivenName
	model.MiddleName = sam.MiddleName
	model.NickName = sam.NickName
	model.PreferredUsername = sam.PreferredUsername
	model.Profile = sam.Profile
	model.Picture = sam.Picture
	model.Website = sam.Website
	model.Gender = sam.Gender
	model.Birthdate = sam.Birthdate
	model.ZoneInfo = sam.ZoneInfo
	model.Locale = sam.Locale
	model.SamlUpdatedAt = sam.UpdatedAt
	model.Email = sam.Email
	model.EmailVerified = sam.EmailVerified
	model.Phone = sam.Phone
	model.PhoneVerified = sam.PhoneVerified

	return model
}

type SamlCreateProviderRequest struct {
	Enabled               bool                                   `json:"enabled" validate:"omitempty,boolean"`
	Name                  string                                 `json:"name" validate:"required,min=5"`
	Domain                string                                 `json:"domain" validate:"required,hostname_rfc1123"`
	MetadataUrl           string                                 `json:"metadata_url" validate:"required,url"`
	SkipEmailVerification bool                                   `json:"skip_email_verification" validate:"omitempty,boolean"`
	AttributeMap          *SamlCreateProviderAttributeMapRequest `json:"attribute_map" validate:"omitempty"`
}

func (s *SamlCreateProviderRequest) ToModel() (*models.SamlIdentityProvider, error) {
	now := time.Now()

	providerId, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("unable to generate uuid: %w", err)
	}

	attributeMapId, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("unable to generate uuid: %w", err)
	}

	attributeMapModel := &models.SamlAttributeMap{
		ID:                 attributeMapId,
		IdentityProviderID: providerId,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if s.AttributeMap != nil {
		attributeMapModel = s.AttributeMap.ToModel(attributeMapModel)
	}

	provider := models.SamlIdentityProvider{
		ID:                    providerId,
		AttributeMap:          *attributeMapModel,
		Enabled:               s.Enabled,
		Name:                  s.Name,
		Domain:                s.Domain,
		MetadataUrl:           s.MetadataUrl,
		SkipEmailVerification: s.SkipEmailVerification,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	return &provider, nil
}

type SamlGetProviderRequest struct {
	ID string `param:"id" validate:"required,uuid4"`
}

type SamlUpdateProviderRequest struct {
	SamlGetProviderRequest
	SamlCreateProviderRequest
}

func (su *SamlUpdateProviderRequest) UpdateModelFromDto(model *models.SamlIdentityProvider) *models.SamlIdentityProvider {
	now := time.Now()

	model.Enabled = su.Enabled
	model.Name = su.Name
	model.Domain = su.Domain
	model.MetadataUrl = su.MetadataUrl
	model.SkipEmailVerification = su.SkipEmailVerification
	model.UpdatedAt = now

	model.AttributeMap.Name = su.AttributeMap.Name
	model.AttributeMap.FamilyName = su.AttributeMap.FamilyName
	model.AttributeMap.GivenName = su.AttributeMap.GivenName
	model.AttributeMap.MiddleName = su.AttributeMap.MiddleName
	model.AttributeMap.NickName = su.AttributeMap.NickName
	model.AttributeMap.PreferredUsername = su.AttributeMap.PreferredUsername
	model.AttributeMap.Profile = su.AttributeMap.Profile
	model.AttributeMap.Picture = su.AttributeMap.Picture
	model.AttributeMap.Website = su.AttributeMap.Website
	model.AttributeMap.Gender = su.AttributeMap.Gender
	model.AttributeMap.Birthdate = su.AttributeMap.Birthdate
	model.AttributeMap.ZoneInfo = su.AttributeMap.ZoneInfo
	model.AttributeMap.Locale = su.AttributeMap.Locale
	model.AttributeMap.SamlUpdatedAt = su.AttributeMap.UpdatedAt
	model.AttributeMap.Email = su.AttributeMap.Email
	model.AttributeMap.EmailVerified = su.AttributeMap.EmailVerified
	model.AttributeMap.Phone = su.AttributeMap.Phone
	model.AttributeMap.PhoneVerified = su.AttributeMap.PhoneVerified

	return model
}
