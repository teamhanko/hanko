package config

type SecurityNotifications struct {
	Notifications SecurityNotificationTypes `yaml:"notifications" json:"notifications,omitempty" koanf:"notifications"`
}

type SecurityNotificationTypes struct {
	PasswordUpdate     SecurityNotificationConfiguration `yaml:"password_update" json:"password_update,omitempty" koanf:"password_update"`
	PrimaryEmailUpdate SecurityNotificationConfiguration `yaml:"primary_email_update" json:"primary_email_update,omitempty" koanf:"primary_email_update"`
	EmailCreate        SecurityNotificationConfiguration `yaml:"email_create" json:"email_create,omitempty" koanf:"email_create"`
	EmailDelete        SecurityNotificationConfiguration `yaml:"email_delete" json:"email_delete,omitempty" koanf:"email_delete"`
	PasskeyCreate      SecurityNotificationConfiguration `yaml:"passkey_create" json:"passkey_create,omitempty" koanf:"passkey_create"`
	MFAEnabled         SecurityNotificationConfiguration `yaml:"mfa_enabled" json:"mfa_enabled,omitempty" koanf:"mfa_enabled"`
	MFADisabled        SecurityNotificationConfiguration `yaml:"mfa_disabled" json:"mfa_disabled,omitempty" koanf:"mfa_disabled"`
}

type SecurityNotificationConfiguration struct {
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
}
