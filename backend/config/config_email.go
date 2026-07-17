package config

import "fmt"

type Email struct {
	// `acquire_on_login` determines whether users, provided that they do not already have registered an email,
	//	are prompted to provide an email on login.
	AcquireOnLogin bool `yaml:"acquire_on_login" json:"acquire_on_login" koanf:"acquire_on_login" split_words:"true" jsonschema:"default=true"`
	// `acquire_on_registration` determines whether users are prompted to provide an email on registration.
	AcquireOnRegistration bool `yaml:"acquire_on_registration" json:"acquire_on_registration" koanf:"acquire_on_registration" split_words:"true" jsonschema:"default=true"`
	// `enabled` determines whether emails are enabled.
	Enabled bool `yaml:"enabled" json:"enabled" koanf:"enabled" jsonschema:"default=true"`
	// 'limit' determines the maximum number of emails a user can register.
	Limit int `yaml:"limit" json:"limit" koanf:"limit" jsonschema:"default=5"`
	// `max_length` specifies the maximum allowed length of an email address.
	MaxLength int `yaml:"max_length" json:"max_length" koanf:"max_length" jsonschema:"default=120"`
	// `optional` determines whether users must provide an email when prompted.
	// There must always be at least one email address associated with an account. The primary email address cannot be
	// deleted if emails are required (`optional`: false`).
	Optional bool `yaml:"optional" json:"optional" koanf:"optional" jsonschema:"default=false"`
	// `passcode_ttl` specifies, in seconds, how long a passcode is valid for.
	PasscodeTtl int `yaml:"passcode_ttl" json:"passcode_ttl" koanf:"passcode_ttl" jsonschema:"default=300"`
	// `passcode_charset` specifies the characters that can be used in passcodes.
	// E.g. `numeric` allows only numbers, `alphanumeric` allows both numbers and letters.
	PasscodeCharset PasscodeCharset `yaml:"passcode_charset" json:"passcode_charset" koanf:"passcode_charset" jsonschema:"default=numeric,enum=numeric,enum=alphanumeric"`
	// `require_verification` determines whether newly created emails must be verified by providing a passcode sent
	// to respective address.
	RequireVerification bool `yaml:"require_verification" json:"require_verification" koanf:"require_verification" split_words:"true" jsonschema:"default=true"`
	// `use_as_login_identifier` determines whether emails can be used as an identifier on login.
	UseAsLoginIdentifier bool `yaml:"use_as_login_identifier" json:"use_as_login_identifier" koanf:"use_as_login_identifier" jsonschema:"default=true"`
	// `user_for_authentication` determines whether users can log in by providing an email address and subsequently
	// providing a passcode sent to the given email address.
	UseForAuthentication bool `yaml:"use_for_authentication" json:"use_for_authentication" koanf:"use_for_authentication" jsonschema:"default=true"`
}

type PasscodeCharset string

var (
	PasscodeCharsetNumeric      PasscodeCharset = "numeric"
	PasscodeCharsetAlphanumeric PasscodeCharset = "alphanumeric"
)

func (e *Email) Validate() error {
	switch e.PasscodeCharset {
	case PasscodeCharsetNumeric, PasscodeCharsetAlphanumeric:
		return nil
	}
	return fmt.Errorf("invalid passcode_characters: %s (allowed: 'numeric', 'alphanumeric')", e.PasscodeCharset)
}

func (e *Email) PostProcess() error {
	if e.PasscodeCharset == "" {
		e.PasscodeCharset = PasscodeCharsetNumeric
	}

	return nil
}
