package config

type SecurityNotifications struct {
	Notifications SecurityNotificationTypes `yaml:"notifications" json:"notifications" koanf:"notifications"`
}

type SecurityNotificationTypes struct {
	PasswordUpdate     SecurityNotificationConfiguration `yaml:"password_update" json:"password_update" koanf:"password_update"`
	PrimaryEmailUpdate SecurityNotificationConfiguration `yaml:"primary_email_update" json:"primary_email_update" koanf:"primary_email_update"`
	EmailCreate        SecurityNotificationConfiguration `yaml:"email_create" json:"email_create" koanf:"email_create"`
	EmailDelete        SecurityNotificationConfiguration `yaml:"email_delete" json:"email_delete" koanf:"email_delete"`
	PasskeyCreate      SecurityNotificationConfiguration `yaml:"passkey_create" json:"passkey_create" koanf:"passkey_create"`
	MFACreate          SecurityNotificationConfiguration `yaml:"mfa_create" json:"mfa_create" koanf:"mfa_create"`
	MFADelete          SecurityNotificationConfiguration `yaml:"mfa_delete" json:"mfa_delete" koanf:"mfa_delete"`
}

type SecurityNotificationConfiguration struct {
	Enabled bool `yaml:"enabled" json:"enabled" koanf:"enabled" jsonschema:"default=true"`
}
