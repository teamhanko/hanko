package config

import "github.com/invopop/jsonschema"

type Password struct {
	// `acquire_on_registration` configures how users are prompted creating a password on registration.
	AcquireOnRegistration string `yaml:"acquire_on_registration" json:"acquire_on_registration,omitempty" koanf:"acquire_on_registration" split_words:"true" jsonschema:"default=always,enum=always,enum=conditional,enum=never"`
	// `acquire_on_login` configures how users are prompted creating a password on login.
	AcquireOnLogin string `yaml:"acquire_on_login" json:"acquire_on_login,omitempty" koanf:"acquire_on_login" split_words:"true" jsonschema:"default=never,enum=always,enum=conditional,enum=never"`
	// `enabled` determines whether passwords are enabled or disabled.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	// `min_length` determines the minimum password length.
	MinLength int `yaml:"min_length" json:"min_length,omitempty" koanf:"min_length" split_words:"true" jsonschema:"default=8"`
	// Deprecated. Use `min_length` instead.
	MinPasswordLength int `yaml:"min_password_length" json:"min_password_length,omitempty" koanf:"min_password_length" split_words:"true" jsonschema:"default=8"`
	// `optional` determines whether users must set a password when prompted. The password cannot be deleted if
	// passwords are required (`optional: false`).
	//
	// It also takes part in determining the order of password and passkey acquisition
	// on login and registration (see also `acquire_on_login` and `acquire_on_registration`): if one credential type is
	// required (`optional: false`) then that one takes precedence, i.e. is acquired first.
	Optional bool `yaml:"optional" json:"optional,omitempty" koanf:"optional" jsonschema:"default=false"`
	// `recovery` determines whether users can start a recovery process, e.g. in case of a forgotten password.
	Recovery bool `yaml:"recovery" json:"recovery,omitempty" koanf:"recovery" jsonschema:"default=true"`
}

func (Password) JSONSchemaExtend(schema *jsonschema.Schema) {
	acquireOnRegistration, _ := schema.Properties.Get("acquire_on_registration")
	acquireOnRegistration.Extras = map[string]any{"meta:enum": map[string]string{
		"always": "Indicates that users are always prompted to create a password on registration.",
		"conditional": `Indicates that users are prompted to create a password on registration as long as the user does
						not have a passkey.

						If passkeys are also conditionally acquired on registration, then users are given a choice as
						to what type of credential to register.`,
		"never": "Indicates that users are never prompted to create a password on registration.",
	}}

	acquireOnLogin, _ := schema.Properties.Get("acquire_on_login")
	acquireOnLogin.Extras = map[string]any{"meta:enum": map[string]string{
		"always": `Indicates that users are always prompted to create a password on login
					provided that they do not already have a password.`,
		"conditional": `Indicates that users are prompted to create a password on login provided that
						they do not already have a password and do not have a passkey.

						If passkeys are also conditionally acquired on login then users are given a choice as to what
						type of credential to register.`,
		"never": "Indicates that users are never prompted to create a password on login.",
	}}
}
