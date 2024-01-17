package webhooks

import (
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	cfg := config.Config{}
	jwkManager := test.JwkManager{}
	persister := test.NewPersister(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	manager, err := NewManager(&cfg, persister.GetWebhookPersister(nil), jwkManager, nil)
	assert.NoError(t, err)
	require.NotEmpty(t, manager)
}

func TestManager_GenerateJWT(t *testing.T) {
	cfg := config.Config{}
	jwkManager := test.JwkManager{}
	persister := test.NewPersister(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	manager, err := NewManager(&cfg, persister.GetWebhookPersister(nil), jwkManager, nil)

	testData := "lorem-ipsum"

	dataToken, err := manager.GenerateJWT(testData, events.User)
	assert.NoError(t, err)
	assert.NotEmpty(t, dataToken)
}

func TestManager_TriggerWithoutHook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "no hook should not trigger a http request")
	}))
	defer server.Close()

	cfg := config.Config{}
	jwkManager := test.JwkManager{}
	persister := test.NewPersister(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	manager, err := NewManager(&cfg, persister.GetWebhookPersister(nil), jwkManager, nil)
	assert.NoError(t, err)

	manager.Trigger(events.User, "lorem-ipsum")

	// give it 1 sec to trigger
	time.Sleep(1 * time.Second)
}
func TestManager_TriggerWithConfigHook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.True(t, true)
	}))
	defer server.Close()

	hooks := config.Webhooks{config.Webhook{
		Callback: server.URL,
		Events: events.Events{
			events.User,
		},
	}}

	cfg := config.Config{
		Webhooks: config.WebhookSettings{
			Enabled: true,
			Hooks:   hooks,
		},
	}

	jwkManager := test.JwkManager{}
	persister := test.NewPersister(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	manager, err := NewManager(&cfg, persister.GetWebhookPersister(nil), jwkManager, nil)
	assert.NoError(t, err)

	manager.Trigger(events.User, "lorem-ipsum")

	// give it 1 sec to trigger
	time.Sleep(1 * time.Second)
}

func TestManager_TriggerWithDisabledConfigHook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "no hook should not trigger a http request")
	}))
	defer server.Close()

	hooks := config.Webhooks{config.Webhook{
		Callback: server.URL,
		Events: events.Events{
			events.User,
		},
	}}

	cfg := config.Config{
		Webhooks: config.WebhookSettings{
			Enabled: false,
			Hooks:   hooks,
		},
	}

	jwkManager := test.JwkManager{}
	persister := test.NewPersister(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	manager, err := NewManager(&cfg, persister.GetWebhookPersister(nil), jwkManager, nil)
	assert.NoError(t, err)

	manager.Trigger(events.User, "lorem-ipsum")

	// give it 1 sec to trigger
	time.Sleep(1 * time.Second)
}

func TestManager_TriggerWithDbHook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.True(t, true)
	}))
	defer server.Close()

	hookUuid, err := uuid.NewV4()
	assert.NoError(t, err)

	eventUuid, err := uuid.NewV4()
	assert.NoError(t, err)

	cfg := config.Config{}
	jwkManager := test.JwkManager{}
	persister := test.NewPersister(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		models.Webhooks{
			models.Webhook{
				ID:        hookUuid,
				Callback:  server.URL,
				Enabled:   true,
				Failures:  0,
				ExpiresAt: time.Now(),
				WebhookEvents: models.WebhookEvents{
					models.WebhookEvent{
						ID:        eventUuid,
						Webhook:   nil,
						WebhookID: hookUuid,
						Event:     string(events.User),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		nil,
	)

	manager, err := NewManager(&cfg, persister.GetWebhookPersister(nil), jwkManager, nil)
	assert.NoError(t, err)

	manager.Trigger(events.User, "lorem-ipsum")

	// give it 1 sec to trigger
	time.Sleep(1 * time.Second)
}

func TestManager_TriggerWithDisabledDbHook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "no hook should not trigger a http request")
	}))
	defer server.Close()

	hookUuid, err := uuid.NewV4()
	assert.NoError(t, err)

	eventUuid, err := uuid.NewV4()
	assert.NoError(t, err)

	cfg := config.Config{}
	jwkManager := test.JwkManager{}
	persister := test.NewPersister(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		models.Webhooks{
			models.Webhook{
				ID:        hookUuid,
				Callback:  server.URL,
				Enabled:   false,
				Failures:  0,
				ExpiresAt: time.Now(),
				WebhookEvents: models.WebhookEvents{
					models.WebhookEvent{
						ID:        eventUuid,
						Webhook:   nil,
						WebhookID: hookUuid,
						Event:     string(events.User),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		nil,
	)

	manager, err := NewManager(&cfg, persister.GetWebhookPersister(nil), jwkManager, nil)
	assert.NoError(t, err)

	manager.Trigger(events.User, "lorem-ipsum")

	// give it 1 sec to trigger
	time.Sleep(1 * time.Second)
}
