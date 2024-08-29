package config

type Passcode struct {
	// Deprecated. Use `email.passcode_ttl` instead.
	TTL int `yaml:"ttl" json:"ttl,omitempty" koanf:"ttl" jsonschema:"default=300"`
}
