package config

import (
	"github.com/invopop/jsonschema"
	"time"
)

type SecurityKeys struct {
	// `attestation_preference` is used to specify the preference regarding attestation conveyance during
	// credential generation.
	AttestationPreference string `yaml:"attestation_preference" json:"attestation_preference,omitempty" koanf:"attestation_preference" split_words:"true" jsonschema:"default=direct,enum=direct,enum=indirect,enum=none"`
	// `authenticator_attachment`  is used to specify the preference regarding authenticator attachment during credential registration.
	AuthenticatorAttachment string `yaml:"authenticator_attachment" json:"authenticator_attachment,omitempty" koanf:"authenticator_attachment" split_words:"true" jsonschema:"default=cross-platform,enum=platform,enum=cross-platform,enum=no_preference"`
	// `enabled` determines whether security keys are eligible for multi-factor-authentication.
	Enabled bool `yaml:"enabled" json:"enabled" koanf:"enabled" jsonschema:"default=true"`
	// 'limit' determines the maximum number of security keys a user can register.
	Limit int `yaml:"limit" json:"limit,omitempty" koanf:"limit" jsonschema:"default=10"`
	// `user_verification` specifies the requirements regarding local authorization with an authenticator through
	//  various authorization gesture modalities; for example, through a touch plus pin code,
	//  password entry, or biometric recognition.
	//
	// The setting applies to both WebAuthn registration and authentication ceremonies.
	UserVerification string `yaml:"user_verification" json:"user_verification,omitempty" koanf:"user_verification" split_words:"true" jsonschema:"default=discouraged,enum=required,enum=preferred,enum=discouraged"`
}

type TOTP struct {
	// `enabled` determines whether TOTP is eligible for multi-factor-authentication.
	Enabled bool `yaml:"enabled" json:"enabled" koanf:"enabled" jsonschema:"default=true"`
}

type MFA struct {
	// `acquire_on_login` configures if users are prompted creating an MFA credential on login.
	AcquireOnLogin bool `yaml:"acquire_on_login" json:"acquire_on_login" koanf:"acquire_on_login" jsonschema:"default=false"`
	// `acquire_on_registration` configures if users are prompted creating an MFA credential on registration.
	AcquireOnRegistration bool `yaml:"acquire_on_registration" json:"acquire_on_registration" koanf:"acquire_on_registration" jsonschema:"default=true"`
	// `device_trust_cookie_name` is the name of the cookie used to store the token of a trusted device.
	DeviceTrustCookieName string `yaml:"device_trust_cookie_name" json:"device_trust_cookie_name,omitempty" koanf:"device_trust_cookie_name" jsonschema:"default=hanko_device_token"`
	// `device_trust_duration` configures the duration a device remains trusted after authentication; once expired, the
	// user must reauthenticate with MFA.
	DeviceTrustDuration time.Duration `yaml:"device_trust_duration" json:"device_trust_duration" koanf:"device_trust_duration" jsonschema:"default=720h,type=string"`
	// `device_trust_policy` determines the conditions under which a device or browser is considered trusted, allowing
	// MFA to be skipped for subsequent logins.
	DeviceTrustPolicy string `yaml:"device_trust_policy" json:"device_trust_policy,omitempty" koanf:"device_trust_policy" split_words:"true" jsonschema:"default=prompt,enum=always,enum=prompt,enum=never"`
	// `enabled` determines whether multi-factor-authentication is enabled.
	Enabled bool `yaml:"enabled" json:"enabled" koanf:"enabled" jsonschema:"default=true"`
	// `optional` determines whether users must create an MFA credential when prompted. The MFA credential cannot be
	// deleted if multi-factor-authentication is required (`optional: false`).
	Optional bool `yaml:"optional" json:"optional" koanf:"optional" jsonschema:"default=true"`
	// `security_keys` configures security key settings for multi-factor-authentication
	SecurityKeys SecurityKeys `yaml:"security_keys" json:"security_keys,omitempty" koanf:"security_keys" jsonschema:"title=security_keys"`
	// `totp` configures the TOTP (Time-Based One-Time-Password) method for multi-factor-authentication.
	TOTP TOTP `yaml:"totp" json:"totp,omitempty" koanf:"totp" jsonschema:"title=totp"`
}

func (MFA) JSONSchemaExtend(schema *jsonschema.Schema) {
	deviceTrustPolicy, _ := schema.Properties.Get("device_trust_policy")
	deviceTrustPolicy.Extras = map[string]any{"meta:enum": map[string]string{
		"always": "Devices are trusted without user consent until the trust expires, so MFA is skipped during subsequent logins.",
		"prompt": "The user can choose to trust the current device to skip MFA for subsequent logins.",
		"never":  "Devices are considered untrusted, so MFA is required for each login.",
	}}
}
