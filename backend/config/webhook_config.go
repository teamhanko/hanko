package config

import (
	"encoding/json"
	"fmt"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"net/url"
	"strings"
)

type WebhookSettings struct {
	Enabled             bool     `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	AllowTimeExpiration bool     `yaml:"allow_time_expiration" json:"allow_time_expiration,omitempty" koanf:"allow_time_expiration" jsonschema:"default=false"`
	Hooks               Webhooks `yaml:"hooks" json:"hooks,omitempty" koanf:"hooks"`
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

// Decode is an implementation of the envconfig.Decoder interface.
// Assumes that environment variables (for the WEBHOOKS_HOOKS key) have the following format:
// {"callback":"http://app.com/usercb","events":["user"]};{"callback":"http://app.com/emailcb","events":["email.send"]}
func (wd *Webhooks) Decode(value string) error {
	webhooks := Webhooks{}
	hooks := strings.Split(value, ";")
	for _, hook := range hooks {
		webhook := Webhook{}
		err := json.Unmarshal([]byte(hook), &webhook)
		if err != nil {
			return fmt.Errorf("invalid map json: %w", err)
		}
		webhooks = append(webhooks, webhook)

	}
	*wd = webhooks
	return nil
}

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
