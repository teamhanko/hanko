package webhooks

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"testing"
	"time"
)

func TestNewConfigHook(t *testing.T) {
	hook := config.Webhook{
		Callback: "http://lorem.ipsum",
		Events:   events.Events{events.User},
	}

	cfgHook := NewConfigHook(hook, nil)
	require.NotEmpty(t, cfgHook)
}

func TestConfigHook_DisableOnExpiryDate(t *testing.T) {
	now := time.Now()
	hook := config.Webhook{
		Callback: "http://lorem.ipsum",
		Events:   events.Events{events.User},
	}

	dbHook := NewConfigHook(hook, nil)
	err := dbHook.DisableOnExpiryDate(now)
	assert.NoError(t, err)
}

func TestConfigHook_DisableOnFailure(t *testing.T) {
	hook := config.Webhook{
		Callback: "http://lorem.ipsum",
		Events:   events.Events{events.User},
	}

	dbHook := NewConfigHook(hook, nil)
	err := dbHook.DisableOnFailure()
	assert.NoError(t, err)
}

func TestConfigHook_Reset(t *testing.T) {
	hook := config.Webhook{
		Callback: "http://lorem.ipsum",
		Events:   events.Events{events.User},
	}

	dbHook := NewConfigHook(hook, nil)
	err := dbHook.Reset()
	assert.NoError(t, err)
}

func TestConfigHook_IsEnabled(t *testing.T) {
	hook := config.Webhook{
		Callback: "http://lorem.ipsum",
		Events:   events.Events{events.User},
	}

	dbHook := NewConfigHook(hook, nil)
	isEnabled := dbHook.IsEnabled()
	require.True(t, isEnabled)
}
