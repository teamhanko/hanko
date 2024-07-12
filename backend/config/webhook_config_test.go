package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
