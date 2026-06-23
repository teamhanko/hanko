package dto

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

func enabledBuiltIn(id string) config.ThirdPartyProvider {
	return config.ThirdPartyProvider{ID: id, Enabled: true}
}

func enabledCustom(id string) config.CustomThirdPartyProvider {
	return config.CustomThirdPartyProvider{ID: id, Enabled: true}
}

func TestIsIdentityForDisabledProvider(t *testing.T) {
	builtIn := []config.ThirdPartyProvider{enabledBuiltIn("google"), enabledBuiltIn("github")}
	custom := []config.CustomThirdPartyProvider{enabledCustom("custom_myoidc")}

	tests := []struct {
		name     string
		identity models.Identity
		want     bool
	}{
		{
			name:     "enabled built-in provider",
			identity: models.Identity{ProviderID: "google"},
			want:     false,
		},
		{
			name:     "disabled built-in provider (not in enabled list)",
			identity: models.Identity{ProviderID: "apple"},
			want:     true,
		},
		{
			name:     "enabled custom provider",
			identity: models.Identity{ProviderID: "custom_myoidc"},
			want:     false,
		},
		{
			name:     "disabled custom provider (not in enabled list)",
			identity: models.Identity{ProviderID: "custom_other"},
			want:     true,
		},
		{
			name: "enabled SAML provider",
			identity: models.Identity{
				SamlIdentity: &models.SamlIdentity{
					SamlProvider: &models.SamlProvider{Enabled: true},
				},
			},
			want: false,
		},
		{
			name: "disabled SAML provider",
			identity: models.Identity{
				SamlIdentity: &models.SamlIdentity{
					SamlProvider: &models.SamlProvider{Enabled: false},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isIdentityForDisabledProvider(tt.identity, builtIn, custom)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFromIdentityModel(t *testing.T) {
	cfg := &config.TenantConfig{
		ThirdParty: config.ThirdParty{
			CustomProviders: config.CustomThirdPartyProviders{
				"myoidc": config.CustomThirdPartyProvider{DisplayName: "My OIDC"},
			},
		},
	}
	id := uuid.Must(uuid.NewV4())

	tests := []struct {
		name     string
		identity *models.Identity
		want     *Identity
	}{
		{
			name:     "nil identity",
			identity: nil,
			want:     nil,
		},
		{
			name: "built-in provider",
			identity: &models.Identity{
				ID:             id,
				ProviderID:     "google",
				ProviderUserID: "user-123",
			},
			want: &Identity{ID: "user-123", Provider: "Google", IdentityID: id},
		},
		{
			name: "custom provider",
			identity: &models.Identity{
				ID:             id,
				ProviderID:     "custom_myoidc",
				ProviderUserID: "user-456",
			},
			want: &Identity{ID: "user-456", Provider: "My OIDC", IdentityID: id},
		},
		{
			name: "SAML provider",
			identity: &models.Identity{
				ID:             id,
				ProviderUserID: "user-789",
				SamlIdentity: &models.SamlIdentity{
					SamlProvider: &models.SamlProvider{Name: "Corp SSO", Enabled: true},
				},
			},
			want: &Identity{ID: "user-789", Provider: "Corp SSO", IdentityID: id},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromIdentityModel(tt.identity, cfg)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetProviderDisplayName(t *testing.T) {
	cfg := &config.TenantConfig{
		ThirdParty: config.ThirdParty{
			CustomProviders: config.CustomThirdPartyProviders{
				"myoidc": config.CustomThirdPartyProvider{DisplayName: "My OIDC"},
			},
		},
	}

	tests := []struct {
		name     string
		identity *models.Identity
		want     string
	}{
		{
			name:     "built-in google",
			identity: &models.Identity{ProviderID: "google"},
			want:     "Google",
		},
		{
			name:     "built-in github",
			identity: &models.Identity{ProviderID: "github"},
			want:     "GitHub",
		},
		{
			name:     "built-in linkedin",
			identity: &models.Identity{ProviderID: "linkedin"},
			want:     "LinkedIn",
		},
		{
			name:     "custom provider",
			identity: &models.Identity{ProviderID: "custom_myoidc"},
			want:     "My OIDC",
		},
		{
			name: "SAML provider",
			identity: &models.Identity{
				SamlIdentity: &models.SamlIdentity{
					SamlProvider: &models.SamlProvider{Name: "Corp SSO", Enabled: true},
				},
			},
			want: "Corp SSO",
		},
		{
			name:     "unknown provider falls back to provider ID",
			identity: &models.Identity{ProviderID: "unknown"},
			want:     "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getProviderDisplayName(tt.identity, cfg)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFromIdentitiesModel_FiltersDisabledProviders(t *testing.T) {
	cfg := &config.TenantConfig{
		ThirdParty: config.ThirdParty{
			Providers: config.ThirdPartyProviders{
				Google: config.ThirdPartyProvider{Enabled: true, ID: "google"},
				Apple:  config.ThirdPartyProvider{Enabled: false, ID: "apple"},
			},
			CustomProviders: config.CustomThirdPartyProviders{
				"myoidc": config.CustomThirdPartyProvider{Enabled: true, ID: "custom_myoidc"},
				"other":  config.CustomThirdPartyProvider{Enabled: false, ID: "custom_other"},
			},
		},
	}

	identities := models.Identities{
		{ProviderID: "google", ProviderUserID: "g1"},
		{ProviderID: "apple", ProviderUserID: "a1"},
		{ProviderID: "custom_myoidc", ProviderUserID: "c1"},
		{ProviderID: "custom_other", ProviderUserID: "c2"},
	}

	result := FromIdentitiesModel(identities, cfg)

	assert.Len(t, result, 2)
	providerIDs := make([]string, len(result))
	for i, r := range result {
		providerIDs[i] = r.ID
	}
	assert.Contains(t, providerIDs, "g1")
	assert.Contains(t, providerIDs, "c1")
}
