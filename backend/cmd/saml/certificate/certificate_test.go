package certificate

import (
	"encoding/json"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

func TestTenantSubject(t *testing.T) {
	tenantID := uuid.Must(uuid.FromString(config.DefaultTenantID))

	fileConfig := func(multiTenancy bool) *config.Config {
		cfg := &config.Config{}
		cfg.MultiTenancy.Enabled = multiTenancy
		cfg.Service.Name = "File App"
		cfg.Saml.Endpoint = "https://file.example.com"
		return cfg
	}

	tests := []struct {
		name         string
		cfg          *config.Config
		tenantConfig string
		wantCN       string
		wantDNS      []string
	}{
		{
			name:         "without multitenancy the tenant row's config is ignored",
			cfg:          fileConfig(false),
			tenantConfig: `{}`,
			wantCN:       "file.example.com",
			wantDNS:      []string{"file.example.com"},
		},
		{
			name:         "with multitenancy the tenant row's config is used, not the file's",
			cfg:          fileConfig(true),
			tenantConfig: `{"service": {"name": "Row App"}, "saml": {"endpoint_url": "https://row.example.com/"}}`,
			wantCN:       "row.example.com",
			wantDNS:      []string{"row.example.com"},
		},
		{
			name:         "with multitenancy and no endpoint the service name is the common name",
			cfg:          fileConfig(true),
			tenantConfig: `{"service": {"name": "Row App"}}`,
			wantCN:       "Row App",
		},
		{
			name:         "with multitenancy and an empty config the tenant ID is the common name",
			cfg:          fileConfig(true),
			tenantConfig: `{}`,
			wantCN:       tenantID.String(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject := tenantSubject(tt.cfg, &models.Tenant{ID: tenantID, Config: json.RawMessage(tt.tenantConfig)})

			assert.Equal(t, tt.wantCN, subject.CommonName)
			assert.Equal(t, tt.wantDNS, subject.DNSNames)
		})
	}
}
