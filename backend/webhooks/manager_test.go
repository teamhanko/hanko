package webhooks

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/test"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
)

func TestManagerSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(managerSuite))
}

type managerSuite struct {
	test.Suite
}

func (s *managerSuite) testLogger() *log.Logger {
	return log.New("test")
}

func (s *managerSuite) testWebhookSecurity() config.WebhookSecurity {
	return config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeInsecure,
		AllowedSchemes: []string{"http", "https"},
	}
}

func (s *managerSuite) TestNewManager() {
	cfg := config.Config{
		Webhooks: config.WebhookSettings{
			Security: s.testWebhookSecurity(),
		},
	}
	jwkManager := test.JwkManager{}

	manager, err := NewManager(&cfg, s.Storage, jwkManager, s.testLogger())
	s.NoError(err)
	s.NotEmpty(manager)
}

func (s *managerSuite) TestManager_GenerateJWT() {
	cfg := config.Config{
		Webhooks: config.WebhookSettings{
			Security: s.testWebhookSecurity(),
		},
	}
	jwkManager := test.JwkManager{}

	manager, err := NewManager(&cfg, s.Storage, jwkManager, s.testLogger())
	s.Require().NoError(err)

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

	cfg := config.Config{
		Webhooks: config.WebhookSettings{
			Security: s.testWebhookSecurity(),
		},
	}
	jwkManager := test.JwkManager{}

	manager, err := NewManager(&cfg, s.Storage, jwkManager, s.testLogger())
	s.Require().NoError(err)

	manager.Trigger(s.Storage.GetConnection(), events.UserCreate, "lorem-ipsum")

	s.Never(func() bool {
		return triggered
	}, 1*time.Second, 50*time.Millisecond)
}

func (s *managerSuite) TestManager_TriggerWithConfigHook() {
	triggered := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		triggered = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	hooks := config.Webhooks{
		{
			Callback: server.URL,
			Events: events.Events{
				events.UserCreate,
			},
		},
	}

	cfg := config.Config{
		Webhooks: config.WebhookSettings{
			Enabled:  true,
			Hooks:    hooks,
			Security: s.testWebhookSecurity(),
		},
	}

	jwkManager := test.JwkManager{}
	manager, err := NewManager(&cfg, s.Storage, jwkManager, s.testLogger())
	s.Require().NoError(err)

	manager.Trigger(s.Storage.GetConnection(), events.UserCreate, "lorem-ipsum")

	s.Eventually(func() bool {
		return triggered
	}, 1*time.Second, 50*time.Millisecond)
}

func (s *managerSuite) TestManager_TriggerWithDisabledConfigHook() {
	triggered := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		triggered = true
	}))
	defer server.Close()

	hooks := config.Webhooks{
		{
			Callback: server.URL,
			Events: events.Events{
				events.UserCreate,
			},
		},
	}

	cfg := config.Config{
		Webhooks: config.WebhookSettings{
			Enabled:  false,
			Hooks:    hooks,
			Security: s.testWebhookSecurity(),
		},
	}

	jwkManager := test.JwkManager{}
	manager, err := NewManager(&cfg, s.Storage, jwkManager, s.testLogger())
	s.Require().NoError(err)

	manager.Trigger(s.Storage.GetConnection(), events.UserCreate, "lorem-ipsum")

	s.Never(func() bool {
		return triggered
	}, 1*time.Second, 50*time.Millisecond)
}

func (s *managerSuite) TestManager_TriggerWithDbHook() {
	triggered := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		triggered = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.Config{
		Webhooks: config.WebhookSettings{
			Security: s.testWebhookSecurity(),
		},
	}
	jwkManager := test.JwkManager{}

	persister := s.Storage.GetWebhookPersister(nil)

	s.createTestDatabaseWebhook(persister, true, server.URL)

	manager, err := NewManager(&cfg, s.Storage, jwkManager, s.testLogger())
	s.Require().NoError(err)

	manager.Trigger(s.Storage.GetConnection(), events.UserCreate, "lorem-ipsum")

	s.Eventually(func() bool {
		return triggered
	}, 1*time.Second, 50*time.Millisecond)
}

func (s *managerSuite) TestManager_TriggerWithDisabledDbHook() {
	triggered := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		triggered = true
	}))
	defer server.Close()

	cfg := config.Config{
		Webhooks: config.WebhookSettings{
			Security: s.testWebhookSecurity(),
		},
	}
	jwkManager := test.JwkManager{}
	persister := s.Storage.GetWebhookPersister(nil)

	s.createTestDatabaseWebhook(persister, false, server.URL)

	manager, err := NewManager(&cfg, s.Storage, jwkManager, s.testLogger())
	s.Require().NoError(err)

	manager.Trigger(s.Storage.GetConnection(), events.UserCreate, "lorem-ipsum")

	s.Never(func() bool {
		return triggered
	}, 1*time.Second, 50*time.Millisecond)
}

func (s *managerSuite) createTestDatabaseWebhook(persister persistence.WebhookPersister, isEnabled bool, callback string) {
	now := time.Now()
	hookID := uuid.FromStringOrNil("8b00da9a-cacf-45ea-b25d-c1ce0f0d7da1")
	err := persister.Create(
		models.Webhook{
			ID:        hookID,
			Callback:  callback,
			Enabled:   isEnabled,
			Failures:  0,
			ExpiresAt: now.Add(WebhookExpireDuration),
			CreatedAt: now,
			UpdatedAt: now,
		},
		models.WebhookEvents{
			{
				ID:        uuid.FromStringOrNil("8b00da9a-cacf-45ea-b25d-c1ce0f0d7da0"),
				WebhookID: hookID,
				Event:     string(events.UserCreate),
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	)
	s.Require().NoError(err)
}
