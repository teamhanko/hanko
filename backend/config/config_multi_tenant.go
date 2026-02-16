package config

// MultiTenant configures multi-tenant mode for Hanko
type MultiTenant struct {
	// Enabled enables multi-tenant mode. When disabled (default), Hanko operates in single-tenant mode
	// with global email/username uniqueness.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`

	// TenantHeader is the HTTP header name used to identify the tenant.
	// The value must be a valid UUID.
	TenantHeader string `yaml:"tenant_header" json:"tenant_header,omitempty" koanf:"tenant_header" split_words:"true" jsonschema:"default=X-Tenant-ID"`

	// AllowGlobalUsers allows users without a tenant_id (backward compatibility).
	// When true, requests without the tenant header will create/access global users.
	// When false, the tenant header is required for all requests.
	AllowGlobalUsers bool `yaml:"allow_global_users" json:"allow_global_users,omitempty" koanf:"allow_global_users" split_words:"true" jsonschema:"default=true"`

	// AutoProvision enables automatic tenant creation when a request includes a tenant ID
	// that doesn't exist yet. The tenant is created with default values.
	// When false, requests with unknown tenant IDs will return a 404 error.
	AutoProvision bool `yaml:"auto_provision" json:"auto_provision,omitempty" koanf:"auto_provision" split_words:"true" jsonschema:"default=true"`
}

func DefaultMultiTenantConfig() MultiTenant {
	return MultiTenant{
		Enabled:          false,
		TenantHeader:     "X-Tenant-ID",
		AllowGlobalUsers: true,
		AutoProvision:    true,
	}
}
