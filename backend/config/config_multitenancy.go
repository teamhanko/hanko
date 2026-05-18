package config

type MultiTenancy struct {
	// `enabled` determines whether multitenancy mode is enabled.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
}
