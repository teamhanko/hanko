package utils

import (
	"testing"

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
			got := IsIdentityForDisabledProvider(tt.identity, builtIn, custom)
			assert.Equal(t, tt.want, got)
		})
	}
}
