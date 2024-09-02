package config

type SecurityKeys struct {
	// `attestation_preference` is used to specify the preference regarding attestation conveyance during
	// credential generation.
	AttestationPreference string `yaml:"attestation_preference" json:"attestation_preference,omitempty" koanf:"attestation_preference" split_words:"true" jsonschema:"default=direct,enum=direct,enum=indirect,enum=none"`
	// `authenticator_attachment`  is used to specify the preference regarding authenticator attachment during credential registration.
	AuthenticatorAttachment string `yaml:"authenticator_attachment" json:"authenticator_attachment,omitempty" koanf:"authenticator_attachment" split_words:"true" jsonschema:"default=cross-platform,enum=platform,enum=cross-platform,enum=no_preference"`
	// `enabled` determines whether security keys are eligible for multi-factor-authentication.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	// 'limit' determines the maximum number of security keys a user can register.
	Limit int `yaml:"limit" json:"limit,omitempty" koanf:"limit" jsonschema:"default=10"`
	// The setting applies to both WebAuthn registration and authentication ceremonies.
	UserVerification string `yaml:"user_verification" json:"user_verification,omitempty" koanf:"user_verification" split_words:"true" jsonschema:"default=discouraged,enum=required,enum=preferred,enum=discouraged"`
}

type TOTP struct {
	// `enabled` determines whether TOTP is eligible for multi-factor-authentication.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
}

type MFA struct {
	// `acquire_on_login` configures if users are prompted creating an MFA credential on login.
	AcquireOnLogin bool `yaml:"acquire_on_login" json:"acquire_on_login,omitempty" koanf:"acquire_on_login" jsonschema:"default=false"`
	// `acquire_on_registration` configures if users are prompted creating an MFA credential on registration.
	AcquireOnRegistration bool `yaml:"acquire_on_registration" json:"acquire_on_registration,omitempty" koanf:"acquire_on_registration" jsonschema:"default=true"`
	// `enabled` determines whether multi-factor-authentication is enabled.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	// `optional` determines whether users must create an MFA credential when prompted. The MFA credential cannot be
	// deleted if multi-factor-authentication is required (`optional: false`).
	Optional bool `yaml:"optional" json:"optional,omitempty" koanf:"optional" jsonschema:"default=true"`
	// `security_keys` configures security key settings for multi-factor-authentication
	SecurityKeys SecurityKeys `yaml:"security_keys" json:"security_keys,omitempty" koanf:"security_keys" jsonschema:"title=security_keys"`
	// `totp` configures the TOTP (Time-Based One-Time-Password) method for multi-factor-authentication.
	TOTP TOTP `yaml:"totp" json:"totp,omitempty" koanf:"totp" jsonschema:"title=totp"`
}
