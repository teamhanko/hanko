package dto

import (
	"github.com/teamhanko/hanko/backend/config"
)

// PublicConfig is the part of the configuration that will be shared with the frontend
type PublicConfig struct {
	Password                 config.Password `json:"password"`
	EmailVerificationEnabled bool            `json:"email_verification_enabled"`
}

// FromConfig Returns a PublicConfig from the Application configuration
func FromConfig(config config.Config) PublicConfig {
	return PublicConfig{
		Password:                 config.Password,
		EmailVerificationEnabled: config.Registration.EmailVerification.Enabled,
	}
}
