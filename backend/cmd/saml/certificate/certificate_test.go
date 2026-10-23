package certificate

import (
	"encoding/json"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		subject, err := certSubjectResolver.Resolve(tenantWithConfig(tenantConfig))

		require.NoError(t, err)
		assert.Equal(t, "file.example.com", subject.CommonName)
		assert.Equal(t, []string{"file.example.com"}, subject.DNSNames)
	}
}

func TestSubjectResolver_WithoutMultiTenancyNeedsEndpointOrServiceName(t *testing.T) {
	certSubjectResolver := newSubjectResolver(false, config.TenantConfig{})

	_, err := certSubjectResolver.Resolve(tenantWithConfig(`{}`))

	assert.Error(t, err)
}

func TestSubjectResolver_WithMultiTenancyUsesTenantConfig(t *testing.T) {
	certSubjectResolver := newSubjectResolver(true, fileConfigWithEndpoint())

	subject, err := certSubjectResolver.Resolve(tenantWithConfig(
		`{"service": {"name": "Row App"}, "saml": {"endpoint_url": "https://row.example.com/"}}`,
	))

	require.NoError(t, err)
	assert.Equal(t, "row.example.com", subject.CommonName)
	assert.Equal(t, []string{"row.example.com"}, subject.DNSNames)
}

// Plain encoding/json can't read the duration strings that are stored in a tenant's config.
func TestSubjectResolver_WithMultiTenancyReadsDurationStrings(t *testing.T) {
	certSubjectResolver := newSubjectResolver(true, config.TenantConfig{})

	subject, err := certSubjectResolver.Resolve(tenantWithConfig(
		`{"service": {"name": "Row App"}, "mfa": {"device_trust_duration": "720h"}}`,
	))

	require.NoError(t, err)
	assert.Equal(t, "Row App", subject.CommonName)
}

func TestSubjectResolver_WithMultiTenancyWithoutEndpointUsesServiceName(t *testing.T) {
	certSubjectResolver := newSubjectResolver(true, fileConfigWithEndpoint())

	subject, err := certSubjectResolver.Resolve(tenantWithConfig(`{"service": {"name": "Row App"}}`))

	require.NoError(t, err)
	assert.Equal(t, "Row App", subject.CommonName)
	assert.Empty(t, subject.DNSNames)
}

// An empty config is read like the server reads it, with the defaults, which include a service name.
func TestSubjectResolver_WithMultiTenancyEmptyConfigUsesDefaultServiceName(t *testing.T) {
	certSubjectResolver := newSubjectResolver(true, config.TenantConfig{})

	subject, err := certSubjectResolver.Resolve(tenantWithConfig(`{}`))

	require.NoError(t, err)
	assert.Equal(t, config.DefaultTenantConfig().Service.Name, subject.CommonName)
	assert.Empty(t, subject.DNSNames)
}

func TestSubjectResolver_WithMultiTenancyErrors(t *testing.T) {
	certSubjectResolver := newSubjectResolver(true, fileConfigWithEndpoint())

	tests := []struct {
		name         string
		tenantConfig string
	}{
		{"neither endpoint nor service name", `{"service": {"name": ""}}`},
		{"malformed JSON", `{"service": `},
		{"wrong type", `{"service": "Row App"}`},
		{"no config", ``},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := certSubjectResolver.Resolve(tenantWithConfig(tt.tenantConfig))

			assert.Error(t, err)
		})
	}
}
