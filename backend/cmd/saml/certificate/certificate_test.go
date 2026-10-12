package certificate

import (
	"encoding/json"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

func tenantWithConfig(tenantConfig string) *models.Tenant {
	return &models.Tenant{ID: uuid.Must(uuid.FromString(config.DefaultTenantID)), Config: json.RawMessage(tenantConfig)}
}

func fileConfigWithEndpoint() config.TenantConfig {
	var fileConfig config.TenantConfig
	fileConfig.Service.Name = "File App"
	fileConfig.Saml.Endpoint = "https://file.example.com"
	return fileConfig
}

func TestSubjectResolver_WithoutMultiTenancyUsesFileConfig(t *testing.T) {
	certSubjectResolver := newSubjectResolver(false, fileConfigWithEndpoint())

	// the default tenant's row only holds an empty config, and even a filled one is ignored
	for _, tenantConfig := range []string{`{}`, `{"saml": {"endpoint_url": "https://row.example.com"}}`} {
		subject := certSubjectResolver.Resolve(tenantWithConfig(tenantConfig))

		assert.Equal(t, "file.example.com", subject.CommonName)
		assert.Equal(t, []string{"file.example.com"}, subject.DNSNames)
	}
}

func TestSubjectResolver_WithMultiTenancyUsesTenantConfig(t *testing.T) {
	certSubjectResolver := newSubjectResolver(true, fileConfigWithEndpoint())

	subject := certSubjectResolver.Resolve(tenantWithConfig(
		`{"service": {"name": "Row App"}, "saml": {"endpoint_url": "https://row.example.com/"}}`,
	))

	assert.Equal(t, "row.example.com", subject.CommonName)
	assert.Equal(t, []string{"row.example.com"}, subject.DNSNames)
}

func TestSubjectResolver_WithMultiTenancyFallbacks(t *testing.T) {
	certSubjectResolver := newSubjectResolver(true, fileConfigWithEndpoint())

	t.Run("no endpoint: the service name is the common name", func(t *testing.T) {
		subject := certSubjectResolver.Resolve(tenantWithConfig(`{"service": {"name": "Row App"}}`))

		assert.Equal(t, "Row App", subject.CommonName)
		assert.Empty(t, subject.DNSNames)
	})

	t.Run("empty config: the tenant ID is the common name", func(t *testing.T) {
		subject := certSubjectResolver.Resolve(tenantWithConfig(`{}`))

		assert.Equal(t, config.DefaultTenantID, subject.CommonName)
		assert.Empty(t, subject.DNSNames)
	})
}
