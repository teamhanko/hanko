package config

type Username struct {
	// `acquire_on_login` determines whether users, provided that they do not already have set a username,
	//	are prompted to provide a username on login.
	AcquireOnLogin bool `yaml:"acquire_on_login" json:"acquire_on_login,omitempty" koanf:"acquire_on_login" split_words:"true" jsonschema:"default=true"`
	// `acquire_on_registration` determines whether users are prompted to provide a username on registration.
	AcquireOnRegistration bool `yaml:"acquire_on_registration" json:"acquire_on_registration,omitempty" koanf:"acquire_on_registration" split_words:"true" jsonschema:"default=true"`
	// `enabled` determines whether users can set a unique username.
	//
	// Usernames can contain letters (a-z,A-Z), numbers (0-9), and underscores.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	// `max_length` specifies the maximum allowed length of a username.
	MaxLength int `yaml:"max_length" json:"max_length,omitempty" koanf:"max_length" jsonschema:"default=32"`
	// `min_length` specifies the minimum length of a username.
	MinLength int `yaml:"min_length" json:"min_length,omitempty" koanf:"min_length" split_words:"true" jsonschema:"default=3"`
	// `optional` determines whether users must provide a username when prompted. The username can only be changed but
	// not deleted if usernames are required (`optional: false`).
	Optional bool `yaml:"optional" json:"optional,omitempty" koanf:"optional" jsonschema:"default=true"`
	// `use_as_login_identifier` determines whether usernames, if enabled, can be used for logging in.
	UseAsLoginIdentifier bool `yaml:"use_as_login_identifier" json:"use_as_login_identifier,omitempty" koanf:"use_as_login_identifier" jsonschema:"default=true"`
}
