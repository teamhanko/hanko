package dto

import (
	"encoding/json"

	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
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

// CreateSamlProviderRequest represents the request body for creating a SAML provider
type CreateSamlProviderRequest struct {
	Name                  string               `json:"name" validate:"required"`
	MetadataURL           string               `json:"metadata_url" validate:"required,url"`
	Domain                string               `json:"domain" validate:"required"`
	Enabled               bool                 `json:"enabled"`
	SkipEmailVerification bool                 `json:"skip_email_verification"`
	AttributeMap          *config.AttributeMap `json:"attribute_map,omitempty"`
}

// UpdateSamlProviderRequest represents the request body for updating a SAML provider
type UpdateSamlProviderRequest struct {
	Name                  string               `json:"name" validate:"required"`
	MetadataURL           string               `json:"metadata_url" validate:"required,url"`
	Domain                string               `json:"domain" validate:"required"`
	Enabled               bool                 `json:"enabled"`
	SkipEmailVerification bool                 `json:"skip_email_verification"`
	AttributeMap          *config.AttributeMap `json:"attribute_map,omitempty"`
}

// SamlProviderResponse represents the response for a SAML provider
type SamlProviderResponse struct {
	ID                    string               `json:"id"`
	TenantID              string               `json:"tenant_id"`
	Name                  string               `json:"name"`
	EntityID              string               `json:"entity_id"`
	MetadataURL           string               `json:"metadata_url"`
	Domain                string               `json:"domain"`
	Enabled               bool                 `json:"enabled"`
	SkipEmailVerification bool                 `json:"skip_email_verification"`
	AttributeMap          *config.AttributeMap `json:"attribute_map,omitempty"`
	CreatedAt             string               `json:"created_at"`
	UpdatedAt             string               `json:"updated_at"`
}

func FromSamlProvider(provider *models.SamlProvider) SamlProviderResponse {
	resp := SamlProviderResponse{
		ID:                    provider.ID.String(),
		TenantID:              provider.TenantID.String(),
		Name:                  provider.Name,
		EntityID:              provider.EntityID,
		MetadataURL:           provider.MetadataURL,
		Domain:                provider.Domain,
		Enabled:               provider.Enabled,
		SkipEmailVerification: provider.SkipEmailVerification,
		CreatedAt:             provider.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:             provider.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Parse attribute map if present
	if provider.AttributeMap != "" {
		var attrMap config.AttributeMap
		err := json.Unmarshal([]byte(provider.AttributeMap), &attrMap)
		if err == nil {
			resp.AttributeMap = &attrMap
		}
	}

	return resp
}

type SamlProviderResponses []SamlProviderResponse

func FromSamlProviders(providers []models.SamlProvider) SamlProviderResponses {
	responses := make(SamlProviderResponses, len(providers))
	for i, provider := range providers {
		responses[i] = FromSamlProvider(&provider)
	}
	return responses
}
