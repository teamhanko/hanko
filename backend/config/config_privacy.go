package config

type Privacy struct {
	// `show_account_existence_hints` determines whether the user should get a user-friendly response rather than a privacy protecting one. E.g. on sign-up, when enabled the user will get "user already exists" response.
	// It only has an effect when emails are enabled.
	ShowAccountExistenceHints bool `yaml:"show_account_existence_hints" json:"show_account_existence_hints,omitempty" koanf:"show_account_existence_hints" split_words:"true" jsonschema:"default=false"`
	// `only_show_actual_login_methods` determines whether the user will only be prompted with his configured login methods.
	// It only has an effect when emails are enabled, can be used for authentication and passwords are enabled.
	OnlyShowActualLoginMethods bool `yaml:"only_show_actual_login_methods" json:"only_show_actual_login_methods,omitempty" koanf:"only_show_actual_login_methods" split_words:"true" jsonschema:"default=false"`
}
