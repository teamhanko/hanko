package config

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"net/url"
)

type WebhookSettings struct {
	Enabled bool     `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	Hooks   Webhooks `yaml:"hooks" json:"hooks,omitempty" koanf:"hooks"`
}

func (ws *WebhookSettings) Validate() error {
	if ws.Enabled {
		for _, hook := range ws.Hooks {
			err := hook.Validate()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type Webhooks []Webhook

type Webhook struct {
	Callback string        `yaml:"callback" json:"callback,omitempty" koanf:"callback"`
	Events   events.Events `yaml:"events" json:"events,omitempty" koanf:"events"`
}

func (w *Webhook) Validate() error {
	_, err := url.Parse(w.Callback)
	if err != nil {
		return fmt.Errorf("callback is not a valid URL: %w", err)
	}

	if len(w.Events) > 0 {
		for i, e := range w.Events {
			isValid := events.IsValidEvent(e)
			if !isValid {
				return fmt.Errorf("event '%s' at events[%d] is not a valid webhook event", e, i)
			}
		}
	}

	return nil
}
