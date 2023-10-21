package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"reflect"
	"testing"
)

func TestDefaultConfigNotEnoughForValidation(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err == nil {
		t.Error("The default config is missing mandatory parameters. This should not validate without error.")
	}
}

func TestDefaultConfigAccountParameters(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, cfg.Account.AllowDeletion, false)
	assert.Equal(t, cfg.Account.AllowSignup, true)
}

func TestDefaultConfigSmtpParameters(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, cfg.Smtp.Port, "465")
}

func TestParseValidConfig(t *testing.T) {
	configPath := "./config.yaml"
	cfg, err := Load(&configPath)
	if err != nil {
		t.Error(err)
	}
	if err := cfg.Validate(); err != nil {
		t.Error(err)
	}
}

func TestMinimalConfigValidates(t *testing.T) {
	configPath := "./minimal-config.yaml"
	cfg, err := Load(&configPath)
	if err != nil {
		t.Error(err)
	}
	if err := cfg.Validate(); err != nil {
		t.Error(err)
	}
}

func TestRateLimiterConfig(t *testing.T) {
	configPath := "./minimal-config.yaml"
	cfg, err := Load(&configPath)

	if err != nil {
		t.Error(err)
	}
	cfg.RateLimiter.Enabled = true
	cfg.RateLimiter.Store = "in_memory"

	if err := cfg.Validate(); err != nil {
		t.Error(err)
	}

	cfg.RateLimiter.Store = "redis"
	if err := cfg.Validate(); err == nil {
		t.Error("when specifying redis, the redis config should also be specified")
	}
	cfg.RateLimiter.Redis = &RedisConfig{
		Address:  "127.0.0.1:9876",
		Password: "password",
	}
	if err := cfg.Validate(); err != nil {
		t.Error(err)
	}

	cfg.RateLimiter.Store = "notvalid"
	if err := cfg.Validate(); err == nil {
		t.Error("notvalid is not a valid backend")
	}
}

func TestEnvironmentVariables(t *testing.T) {
	err := os.Setenv("SMTP_HOST", "valueFromEnvVars")
	require.NoError(t, err)

	err = os.Setenv("WEBAUTHN_RELYING_PARTY_ORIGINS", "https://hanko.io,https://auth.hanko.io")
	require.NoError(t, err)

	configPath := "./minimal-config.yaml"
	cfg, err := Load(&configPath)
	require.NoError(t, err)

	assert.Equal(t, "valueFromEnvVars", cfg.Smtp.Host)
	assert.True(t, reflect.DeepEqual([]string{"https://hanko.io", "https://auth.hanko.io"}, cfg.Webauthn.RelyingParty.Origins))
}
