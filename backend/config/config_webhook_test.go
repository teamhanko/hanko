package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/v3/webhooks/events"
)

func TestWebhooks_Decode(t *testing.T) {
	webhooks := Webhooks{}
	value := "{\"callback\":\"http://app.com/usercb\",\"events\":[\"user\"]};{\"callback\":\"http://app.com/callback\",\"events\":[\"email.send\"]}"
	err := webhooks.Decode(value)

	assert.NoError(t, err)
	assert.Len(t, webhooks, 2, "has 2 elements")
	for _, webhook := range webhooks {
		assert.IsType(t, Webhook{}, webhook)
	}
}

func TestWebhook_Validate_SessionEvents(t *testing.T) {
	webhook := Webhook{
		Callback: "http://app.com/sessioncb",
		Events: events.Events{
			events.Session,
			events.SessionCreate,
			events.SessionCreateFlow,
			events.SessionCreateAdmin,
			events.SessionDelete,
			events.SessionDeleteExplicit,
			events.SessionDeleteExplicitLogout,
			events.SessionDeleteExplicitRevoke,
			events.SessionDeleteAdmin,
			events.SessionDeleteAdminRevoke,
			events.SessionDeletePassive,
			events.SessionDeletePassiveExpire,
			events.SessionDeletePassiveLimit,
		},
	}

	err := webhook.Validate()

	assert.NoError(t, err)
}
