package config

import (
	"encoding/json"
	"fmt"
	"github.com/invopop/jsonschema"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"net/url"
	"strings"
)

type WebhookSettings struct {
	// `allow_time_expiration` determines whether webhooks are disabled when unused for 30 days
	// (only for database webhooks).
	AllowTimeExpiration bool `yaml:"allow_time_expiration" json:"allow_time_expiration,omitempty" koanf:"allow_time_expiration" jsonschema:"default=false"`
	// `enabled` enables the webhook feature.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	// `hooks` is a list of Webhook configurations.
	//
	// When using environment variables the value for the `WEBHOOKS_HOOKS` key must be specified in the following
	// format:
	// `{"callback":"http://app.com/usercb","events":["user"]};{"callback":"http://app.com/emailcb","events":["email.send"]}`
	Hooks Webhooks `yaml:"hooks" json:"hooks,omitempty" koanf:"hooks" jsonschema:"title=hooks"`
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
	// `callback` specifies the URL to which the change data will be sent.
	Callback string `yaml:"callback" json:"callback,omitempty" koanf:"callback"`
	// `events` is a list of events this hook listens for.
	Events events.Events `yaml:"events" json:"events,omitempty" koanf:"events" jsonschema:"title=events"`
}

func (Webhook) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Title = "hooks"
	evts, _ := schema.Properties.Get("events")

	// If the jsonschema.Reflector is configured with the DoNotReference option set to true, then the items property
	// in the schema is nil, hence we simply create a jsonschema.Schema manually, otherwise we'd get a nil pointer
	// exception.
	if evts.Items == nil {
		evts.Items = &jsonschema.Schema{Type: "string"}
	}
	evts.Items.Title = "events"
	evts.Items.Enum = []any{
		"user",
		"user.create",
		"user.delete",
		"user.update",
		"user.update.email",
		"user.update.email.create",
		"user.update.email.delete",
		"user.update.email.primary",
		"email.send",
	}
	evts.Items.Extras = map[string]any{"meta:enum": map[string]string{
		"user":                      "Triggers on: user creation, user deletion, user update, email creation, email deletion, change of primary email",
		"user.create":               "Triggers on: user creation",
		"user.delete":               "Triggers on: user deletion",
		"user.update":               "Triggers on: user update, email creation, email deletion, change of primary email",
		"user.update.email":         "Triggers on: email creation, email deletion, change of primary email",
		"user.update.email.create":  "Triggers on: email creation",
		"user.update.email.delete":  "Triggers on: email deletion",
		"user.update.email.primary": "Triggers on: change of primary email",
		"email.send":                "Triggers on: an email was sent or should be sent",
	}}
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
