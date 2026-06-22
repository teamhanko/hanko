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
	MFACreate          SecurityNotificationConfiguration `yaml:"mfa_create" json:"mfa_create,omitempty" koanf:"mfa_create"`
	MFADelete          SecurityNotificationConfiguration `yaml:"mfa_delete" json:"mfa_delete,omitempty" koanf:"mfa_delete"`
}

type SecurityNotificationConfiguration struct {
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
}
