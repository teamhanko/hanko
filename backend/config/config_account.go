package config

type Account struct {
	// `allow_deletion` determines whether users can delete their accounts.
	AllowDeletion bool `yaml:"allow_deletion" json:"allow_deletion,omitempty" koanf:"allow_deletion" jsonschema:"default=false"`
	// `allow_signup` determines whether users are able to create new accounts.
	AllowSignup bool `yaml:"allow_signup" json:"allow_signup,omitempty" koanf:"allow_signup" jsonschema:"default=true"`
}
