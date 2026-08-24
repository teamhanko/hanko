package webhooks

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
	"github.com/teamhanko/hanko/backend/v3/webhooks/events"
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

	manager, err := NewManager(cfg.TenantConfig, s.Storage, jwkManager, nil)
	s.NoError(err)
	s.NotEmpty(manager)
}

func (s *managerSuite) TestManager_GenerateJWT() {
	cfg := config.Config{}
	jwkManager := test.JwkManager{}

	manager, err := NewManager(cfg.TenantConfig, s.Storage, jwkManager, nil)

	testData := "lorem-ipsum"

	dataToken, err := manager.GenerateJWT(testData, events.UserCreate, uuid.FromStringOrNil(config.DefaultTenantID))
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

	manager, err := NewManager(cfg.TenantConfig, s.Storage, jwkManager, nil)
	s.Require().NoError(err)

	manager.Trigger(s.Storage.GetConnection(), events.UserCreate, "lorem-ipsum", uuid.FromStringOrNil(config.DefaultTenantID))

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
		TenantConfig: config.TenantConfig{
			Webhooks: config.WebhookSettings{
				Enabled: true,
				Hooks:   hooks,
			},
		},
	}

	jwkManager := test.JwkManager{}
	manager, err := NewManager(cfg.TenantConfig, s.Storage, jwkManager, nil)
	s.Require().NoError(err)

	manager.Trigger(s.Storage.GetConnection(), events.UserCreate, "lorem-ipsum", uuid.FromStringOrNil(config.DefaultTenantID))

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
		TenantConfig: config.TenantConfig{
			Webhooks: config.WebhookSettings{
				Enabled: false,
				Hooks:   hooks,
			},
		},
	}

	jwkManager := test.JwkManager{}
	manager, err := NewManager(cfg.TenantConfig, s.Storage, jwkManager, nil)
	s.Require().NoError(err)

	manager.Trigger(s.Storage.GetConnection(), events.UserCreate, "lorem-ipsum", uuid.FromStringOrNil(config.DefaultTenantID))

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

	manager, err := NewManager(cfg.TenantConfig, s.Storage, jwkManager, nil)
	s.Require().NoError(err)

	manager.Trigger(s.Storage.GetConnection(), events.UserCreate, "lorem-ipsum", uuid.FromStringOrNil(config.DefaultTenantID))

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

	manager, err := NewManager(cfg.TenantConfig, s.Storage, jwkManager, nil)
	s.Require().NoError(err)

	manager.Trigger(s.Storage.GetConnection(), events.UserCreate, "lorem-ipsum", uuid.FromStringOrNil(config.DefaultTenantID))

	// give it 1 sec to trigger
	time.Sleep(1 * time.Second)

	s.False(triggered)
}

func (s *managerSuite) TestManager_TriggerSessionEventHierarchy() {
	tests := []struct {
		name            string
		hookEvent       events.Event
		firedEvent      events.Event
		expectTriggered bool
	}{
		{"session umbrella catches flow create", events.Session, events.SessionCreateFlow, true},
		{"session umbrella catches admin create", events.Session, events.SessionCreateAdmin, true},
		{"session umbrella catches passive delete", events.Session, events.SessionDeletePassiveExpire, true},

		{"session.create catches flow create", events.SessionCreate, events.SessionCreateFlow, true},
		{"session.create catches admin create", events.SessionCreate, events.SessionCreateAdmin, true},
		{"session.create does not catch session delete", events.SessionCreate, events.SessionDeleteExplicitLogout, false},
		{"session.create.flow does not catch admin create", events.SessionCreateFlow, events.SessionCreateAdmin, false},
		{"session.create.admin does not catch flow create", events.SessionCreateAdmin, events.SessionCreateFlow, false},

		{"session.delete catches explicit logout", events.SessionDelete, events.SessionDeleteExplicitLogout, true},
		{"session.delete catches explicit revoke", events.SessionDelete, events.SessionDeleteExplicitRevoke, true},
		{"session.delete catches admin revoke", events.SessionDelete, events.SessionDeleteAdminRevoke, true},
		{"session.delete catches passive expire", events.SessionDelete, events.SessionDeletePassiveExpire, true},
		{"session.delete catches passive limit", events.SessionDelete, events.SessionDeletePassiveLimit, true},

		{"session.delete.explicit catches logout", events.SessionDeleteExplicit, events.SessionDeleteExplicitLogout, true},
		{"session.delete.explicit catches revoke", events.SessionDeleteExplicit, events.SessionDeleteExplicitRevoke, true},
		{"session.delete.explicit does not catch admin revoke", events.SessionDeleteExplicit, events.SessionDeleteAdminRevoke, false},
		{"session.delete.explicit does not catch passive expire", events.SessionDeleteExplicit, events.SessionDeletePassiveExpire, false},

		{"session.delete.admin catches admin revoke", events.SessionDeleteAdmin, events.SessionDeleteAdminRevoke, true},
		{"session.delete.admin does not catch explicit logout", events.SessionDeleteAdmin, events.SessionDeleteExplicitLogout, false},
		{"session.delete.admin does not catch passive expire", events.SessionDeleteAdmin, events.SessionDeletePassiveExpire, false},

		{"session.delete.passive catches expire", events.SessionDeletePassive, events.SessionDeletePassiveExpire, true},
		{"session.delete.passive catches limit", events.SessionDeletePassive, events.SessionDeletePassiveLimit, true},
		{"session.delete.passive does not catch explicit revoke", events.SessionDeletePassive, events.SessionDeleteExplicitRevoke, false},
		{"session.delete.passive does not catch admin revoke", events.SessionDeletePassive, events.SessionDeleteAdminRevoke, false},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			triggered := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				triggered = true
			}))
			defer server.Close()

			hooks := config.Webhooks{config.Webhook{
				Callback: server.URL,
				Events: events.Events{
					currentTest.hookEvent,
				},
			}}

			cfg := config.Config{
				TenantConfig: config.TenantConfig{
					Webhooks: config.WebhookSettings{
						Enabled: true,
						Hooks:   hooks,
					},
				},
			}

			jwkManager := test.JwkManager{}
			manager, err := NewManager(cfg.TenantConfig, s.Storage, jwkManager, nil)
			s.Require().NoError(err)

			manager.Trigger(s.Storage.GetConnection(), currentTest.firedEvent, "lorem-ipsum", uuid.FromStringOrNil(config.DefaultTenantID))

			// give it 1 sec to trigger
			time.Sleep(1 * time.Second)

			s.Equal(currentTest.expectTriggered, triggered)
		})
	}
}

func (s *managerSuite) createTestDatabaseWebhook(persister persistence.WebhookPersister, isEnabled bool, callback string) {
	now := time.Now()
	hookId := uuid.FromStringOrNil("8b00da9a-cacf-45ea-b25d-c1ce0f0d7da1")
	err := persister.Create(
		models.Webhook{
			ID:        hookId,
			TenantID:  uuid.FromStringOrNil(config.DefaultTenantID),
			Callback:  callback,
			Enabled:   isEnabled,
			Failures:  0,
			ExpiresAt: now.Add(WebhookExpireDuration),
			CreatedAt: now,
			UpdatedAt: now,
		},
		models.WebhookEvents{
			models.WebhookEvent{
				ID:        uuid.FromStringOrNil("8b00da9a-cacf-45ea-b25d-c1ce0f0d7da0"),
				TenantID:  uuid.FromStringOrNil(config.DefaultTenantID),
				WebhookID: hookId,
				Event:     string(events.UserCreate),
				CreatedAt: now,
				UpdatedAt: now,
			},
		})
	s.Require().NoError(err)
}
