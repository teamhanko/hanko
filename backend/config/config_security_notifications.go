package config

type SecurityNotifications struct {
	Notifications SecurityNotificationTypes         `yaml:"notifications" json:"notifications,omitempty" koanf:"notifications"`
	Sender        SecurityNotificationsEmailAddress `yaml:"sender" json:"sender,omitempty" koanf:"sender"`
}

type SecurityNotificationTypes struct {
	PasswordUpdate     SecurityNotificationConfiguration `yaml:"password_update" json:"password_update,omitempty" koanf:"password_update"`
	PrimaryEmailUpdate SecurityNotificationConfiguration `yaml:"primary_email_update" json:"primary_email_update,omitempty" koanf:"primary_email_update"`
	EmailCreate        SecurityNotificationConfiguration `yaml:"email_create" json:"email_create,omitempty" koanf:"email_create"`
	PasskeyCreate      SecurityNotificationConfiguration `yaml:"passkey_create" json:"passkey_create,omitempty" koanf:"passkey_create"`
}

type SecurityNotificationConfiguration struct {
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
}

type SecurityNotificationsEmailAddress struct {
	// `from_address` configures the sender address of emails sent to users.
	FromAddress string `yaml:"from_address" json:"from_address,omitempty" koanf:"from_address" split_words:"true" jsonschema:"default=noreply@hanko.io"`

	// `from_name` configures the sender name of emails sent to users.
	FromName string `yaml:"from_name" json:"from_name,omitempty" koanf:"from_name" split_words:"true" jsonschema:"default=Hanko"`
}
