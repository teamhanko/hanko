package config

type Emails struct {
	// Deprecated. Use `email.require_verification` instead.
	RequireVerification bool `yaml:"require_verification" json:"require_verification,omitempty" koanf:"require_verification" split_words:"true" jsonschema:"default=true"`
	// Deprecated. Use `email.limit` instead.
	MaxNumOfAddresses int `yaml:"max_num_of_addresses" json:"max_num_of_addresses,omitempty" koanf:"max_num_of_addresses" split_words:"true" jsonschema:"default=5"`
}
