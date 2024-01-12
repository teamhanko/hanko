package webhooks

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"time"
)

type ConfigHook struct {
	BaseWebhook
}

func NewConfigHook(cfgHook config.Webhook, logger echo.Logger) Webhook {
	return &ConfigHook{
		BaseWebhook{
			Logger:   logger,
			Callback: cfgHook.Callback,
			Events:   cfgHook.Events,
		},
	}
}

func (ch *ConfigHook) DisableOnExpiryDate(_ time.Time) error {
	return nil
}

func (ch *ConfigHook) DisableOnFailure() error {
	return nil
}

func (ch *ConfigHook) Reset() error {
	return nil
}

func (ch *ConfigHook) IsEnabled() bool {
	return false
}
