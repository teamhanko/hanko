package config

import (
	"errors"
	"strings"
)

type EmailDelivery struct {
	// `enabled` determines whether the API delivers emails.
	// Disable if you want to send the emails yourself. To do so you must subscribe to the `email.create` webhook event.
	Enabled bool `yaml:"enabled" json:"enabled" koanf:"enabled" jsonschema:"default=true"`
	// `from_address` configures the sender address of emails sent to users.
	FromAddress string `yaml:"from_address" json:"from_address" koanf:"from_address" split_words:"true" jsonschema:"default=noreply@hanko.io"`
	// `from_name` configures the sender name of emails sent to users.
	FromName string `yaml:"from_name" json:"from_name" koanf:"from_name" split_words:"true" jsonschema:"default=Hanko"`
	// `SMTP` contains the SMTP server settings for sending mails.
	SMTP SMTP `yaml:"smtp" json:"smtp" koanf:"smtp" jsonschema:"title=smtp"`
}

// SMTP Server Settings for sending passcodes
type SMTP struct {
	Host     string `yaml:"host" json:"host" koanf:"host" jsonschema:"default=localhost"`
	Port     string `yaml:"port" json:"port" koanf:"port" jsonschema:"default=465"`
	User     string `yaml:"user" json:"user" koanf:"user"`
	Password string `yaml:"password" json:"password" koanf:"password"`
}

func (s *SMTP) Validate() error {
	if len(strings.TrimSpace(s.Host)) == 0 {
		return errors.New("smtp host must not be empty")
	}
	if len(strings.TrimSpace(s.Port)) == 0 {
		return errors.New("smtp port must not be empty")
	}
	return nil
}
