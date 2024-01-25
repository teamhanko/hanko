package webhooks

import (
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestManagerSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(managerSuite))
}

type managerSuite struct {
	test.Suite
}

func (s *managerSuite) TestNewManager() {
	cfg := config.Config{}
	jwkManager := test.JwkManager{}

	manager, err := NewManager(&cfg, s.Storage.GetWebhookPersister(nil), jwkManager, nil)
	s.NoError(err)
	s.NotEmpty(manager)
}

func (s *managerSuite) TestManager_GenerateJWT() {
	cfg := config.Config{}
	jwkManager := test.JwkManager{}

	manager, err := NewManager(&cfg, s.Storage.GetWebhookPersister(nil), jwkManager, nil)

	testData := "lorem-ipsum"

	dataToken, err := manager.GenerateJWT(testData, events.UserCreate)
	s.NoError(err)
	s.NotEmpty(dataToken)
}

func (s *managerSuite) TestManager_TriggerWithoutHook() {
	triggered := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		triggered = true
	}))
	defer server.Close()

	cfg := config.Config{}
	jwkManager := test.JwkManager{}

	manager, err := NewManager(&cfg, s.Storage.GetWebhookPersister(nil), jwkManager, nil)
	s.Require().NoError(err)

	manager.Trigger(events.UserCreate, "lorem-ipsum")

	// give it 1 sec to trigger
	time.Sleep(1 * time.Second)

	s.False(triggered)
}
func (s *managerSuite) TestManager_TriggerWithConfigHook() {
	triggered := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		triggered = true
	}))
	defer server.Close()

	hooks := config.Webhooks{config.Webhook{
		Callback: server.URL,
		Events: events.Events{
			events.UserCreate,
		},
	}}

	cfg := config.Config{
		Webhooks: config.WebhookSettings{
			Enabled: true,
			Hooks:   hooks,
		},
	}

	jwkManager := test.JwkManager{}
	manager, err := NewManager(&cfg, s.Storage.GetWebhookPersister(nil), jwkManager, nil)
	s.Require().NoError(err)

	manager.Trigger(events.UserCreate, "lorem-ipsum")

	// give it 1 sec to trigger
	time.Sleep(1 * time.Second)

	s.True(triggered)
}

func (s *managerSuite) TestManager_TriggerWithDisabledConfigHook() {
	triggered := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		triggered = true
	}))
	defer server.Close()

	hooks := config.Webhooks{config.Webhook{
		Callback: server.URL,
		Events: events.Events{
			events.UserCreate,
		},
	}}

	cfg := config.Config{
		Webhooks: config.WebhookSettings{
			Enabled: false,
			Hooks:   hooks,
		},
	}

	jwkManager := test.JwkManager{}
	manager, err := NewManager(&cfg, s.Storage.GetWebhookPersister(nil), jwkManager, nil)
	s.Require().NoError(err)

	manager.Trigger(events.UserCreate, "lorem-ipsum")

	// give it 1 sec to trigger
	time.Sleep(1 * time.Second)

	s.False(triggered)
}

func (s *managerSuite) TestManager_TriggerWithDbHook() {
	triggered := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		triggered = true
	}))
	defer server.Close()

	cfg := config.Config{}
	jwkManager := test.JwkManager{}

	persister := s.Storage.GetWebhookPersister(nil)

	s.createTestDatabaseWebhook(persister, true, server.URL)

	manager, err := NewManager(&cfg, persister, jwkManager, nil)
	s.Require().NoError(err)

	manager.Trigger(events.UserCreate, "lorem-ipsum")

	// give it 1 sec to trigger
	time.Sleep(1 * time.Second)

	s.True(triggered)
}

func (s *managerSuite) TestManager_TriggerWithDisabledDbHook() {
	triggered := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		triggered = true
	}))
	defer server.Close()

	cfg := config.Config{}
	jwkManager := test.JwkManager{}
	persister := s.Storage.GetWebhookPersister(nil)

	s.createTestDatabaseWebhook(persister, false, server.URL)

	manager, err := NewManager(&cfg, persister, jwkManager, nil)
	s.Require().NoError(err)

	manager.Trigger(events.UserCreate, "lorem-ipsum")

	// give it 1 sec to trigger
	time.Sleep(1 * time.Second)

	s.False(triggered)
}

func (s *managerSuite) createTestDatabaseWebhook(persister persistence.WebhookPersister, isEnabled bool, callback string) {
	now := time.Now()
	hookId := uuid.FromStringOrNil("8b00da9a-cacf-45ea-b25d-c1ce0f0d7da1")
	err := persister.Create(
		models.Webhook{
			ID:        hookId,
			Callback:  callback,
			Enabled:   isEnabled,
			Failures:  0,
			ExpiresAt: now,
			CreatedAt: now,
			UpdatedAt: now,
		},
		models.WebhookEvents{
			models.WebhookEvent{
				ID:        uuid.FromStringOrNil("8b00da9a-cacf-45ea-b25d-c1ce0f0d7da0"),
				WebhookID: hookId,
				Event:     string(events.UserCreate),
				CreatedAt: now,
				UpdatedAt: now,
			},
		})
	s.Require().NoError(err)
}
