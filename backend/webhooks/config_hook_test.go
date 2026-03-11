package webhooks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
)

func testWebhookSecurity() config.WebhookSecurity {
	return config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeInsecure,
		AllowedSchemes: []string{"http", "https"},
	}
}

func TestNewConfigHook(t *testing.T) {
	hook := config.Webhook{
		Callback: "http://lorem.ipsum",
		Events:   events.Events{events.UserCreate},
	}

	cfgHook := NewConfigHook(hook, testWebhookSecurity(), nil)
	require.NotEmpty(t, cfgHook)
}

func TestConfigHook_DisableOnExpiryDate(t *testing.T) {
	now := time.Now()
	hook := config.Webhook{
		Callback: "http://lorem.ipsum",
		Events:   events.Events{events.UserCreate},
	}

	cfgHook := NewConfigHook(hook, testWebhookSecurity(), nil)
	err := cfgHook.DisableOnExpiryDate(now)
	assert.NoError(t, err)
}

func TestConfigHook_DisableOnFailure(t *testing.T) {
	hook := config.Webhook{
		Callback: "http://lorem.ipsum",
		Events:   events.Events{events.UserCreate},
	}

	cfgHook := NewConfigHook(hook, testWebhookSecurity(), nil)
	err := cfgHook.DisableOnFailure()
	assert.NoError(t, err)
}

func TestConfigHook_Reset(t *testing.T) {
	hook := config.Webhook{
		Callback: "http://lorem.ipsum",
		Events:   events.Events{events.UserCreate},
	}

	cfgHook := NewConfigHook(hook, testWebhookSecurity(), nil)
	err := cfgHook.Reset()
	assert.NoError(t, err)
}

func TestConfigHook_IsEnabled(t *testing.T) {
	hook := config.Webhook{
		Callback: "http://lorem.ipsum",
		Events:   events.Events{events.UserCreate},
	}

	cfgHook := NewConfigHook(hook, testWebhookSecurity(), nil)
	isEnabled := cfgHook.IsEnabled()
	require.True(t, isEnabled)
}
