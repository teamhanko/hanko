package config

type Email struct {
	// `acquire_on_login` determines whether users, provided that they do not already have registered an email,
	//	are prompted to provide an email on login.
	AcquireOnLogin bool `yaml:"acquire_on_login" json:"acquire_on_login,omitempty" koanf:"acquire_on_login" split_words:"true" jsonschema:"default=false"`
	// `acquire_on_registration` determines whether users are prompted to provide an email on registration.
	AcquireOnRegistration bool `yaml:"acquire_on_registration" json:"acquire_on_registration,omitempty" koanf:"acquire_on_registration" split_words:"true" jsonschema:"default=true"`
	// `enabled` determines whether emails are enabled.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	// 'limit' determines the maximum number of emails a user can register.
	Limit int `yaml:"limit" json:"limit,omitempty" koanf:"limit" jsonschema:"default=5"`
	// `max_length` specifies the maximum allowed length of an email address.
	MaxLength int `yaml:"max_length" json:"max_length,omitempty" koanf:"max_length" jsonschema:"default=100"`
	// `optional` determines whether users must provide an email when prompted.
	// There must always be at least one email address associated with an account. The primary email address cannot be
	// deleted if emails are required (`optional`: false`).
	Optional bool `yaml:"optional" json:"optional,omitempty" koanf:"optional" jsonschema:"default=false"`
	// `passcode_ttl` specifies, in seconds, how long a passcode is valid for.
	PasscodeTtl int `yaml:"passcode_ttl" json:"passcode_ttl,omitempty" koanf:"passcode_ttl" jsonschema:"default=300"`
	// `passlink_ttl` specifies, in seconds, how long a passlink is valid for.
	PasslinkTtl int `yaml:"passlink_ttl" json:"passlink_ttl,omitempty" koanf:"passlink_ttl" jsonschema:"default=300"`
	// `require_verification` determines whether newly created emails must be verified by providing a passcode sent
	// to respective address.
	RequireVerification bool `yaml:"require_verification" json:"require_verification,omitempty" koanf:"require_verification" split_words:"true" jsonschema:"default=true"`
	// `use_as_login_identifier` determines whether emails can be used as an identifier on login.
	UseAsLoginIdentifier bool `yaml:"use_as_login_identifier" json:"use_as_login_identifier,omitempty" koanf:"use_as_login_identifier" jsonschema:"default=true"`
	// `user_for_authentication` determines whether users can log in by providing an email address and subsequently
	// providing a passcode sent to the given email address.
	UseForAuthentication bool `yaml:"use_for_authentication" json:"use_for_authentication,omitempty" koanf:"use_for_authentication" jsonschema:"default=true"`
}
