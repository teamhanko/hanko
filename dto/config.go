package dto

import "github.com/teamhanko/hanko/config"

// PublicConfig is the part of the configuration that will be shared with the frontend
type PublicConfig struct {
	Password config.Password `json:"password"`
}

// FromConfig Returns a PublicConfig from the Application configuration
func FromConfig(config config.Config) PublicConfig {
	return PublicConfig{Password: config.Password}
}
