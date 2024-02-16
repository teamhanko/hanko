package webhooks

import (
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"testing"
	"time"
)

func TestDatabaseHookSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(databaseHookSuite))
}

type databaseHookSuite struct {
	test.Suite
}

func (s *databaseHookSuite) TestNewDatabaseHook() {
	hookId, err := uuid.NewV4()
	s.Require().NoError(err)

	hook := models.Webhook{
		ID:        hookId,
		Enabled:   false,
		Failures:  0,
		ExpiresAt: time.Now().Add(WebhookExpireDuration),
	}

	dbHook := NewDatabaseHook(hook, s.Storage.GetWebhookPersister(nil), nil)
	s.NotEmpty(dbHook)
}

func (s *databaseHookSuite) TestDatabaseHook_DisableOnExpiryDate() {
	hook, whPersister := s.loadWebhook("8b00da9a-cacf-45ea-b25d-c1ce0f0d7da3")
	dbHook := NewDatabaseHook(hook, whPersister, nil)

	now := time.Now()
	err := dbHook.DisableOnExpiryDate(now)
	s.NoError(err)

	updatedHook, err := whPersister.Get(hook.ID)
	s.Require().NoError(err)

	s.False(updatedHook.Enabled)
}
func (s *databaseHookSuite) TestDatabaseHook_DoNotDisableOnExpiryDate() {
	hook, whPersister := s.loadWebhook("a47fe92a-1e4b-4119-8653-55ad82737c88")

	dbHook := NewDatabaseHook(hook, whPersister, nil)

	now := time.Now()
	err := dbHook.DisableOnExpiryDate(now)
	s.NoError(err)

	updatedHook, err := whPersister.Get(hook.ID)
	s.Require().NoError(err)

	s.True(updatedHook.Enabled)
}

func (s *databaseHookSuite) TestDatabaseHook_DisableOnFailure() {
	hook, whPersister := s.loadWebhook("8b00da9a-cacf-45ea-b25d-c1ce0f0d7da2")

	dbHook := NewDatabaseHook(hook, whPersister, nil)
	err := dbHook.DisableOnFailure()
	s.Require().NoError(err)

	updatedHook, err := whPersister.Get(hook.ID)
	s.NoError(err)

	s.False(updatedHook.Enabled)
}

func (s *databaseHookSuite) TestDatabaseHook_DoNotDisableOnFailure() {
	hook, whPersister := s.loadWebhook("8b00da9a-cacf-45ea-b25d-c1ce0f0d7da3")

	dbHook := NewDatabaseHook(hook, whPersister, nil)
	err := dbHook.DisableOnFailure()
	s.NoError(err)

	updatedHook, err := whPersister.Get(hook.ID)
	s.Require().NoError(err)

	s.True(updatedHook.Enabled)
}

func (s *databaseHookSuite) TestDatabaseHook_Reset() {
	hook, whPersister := s.loadWebhook("8b00da9a-cacf-45ea-b25d-c1ce0f0d7da2")

	now := time.Now()

	dbHook := NewDatabaseHook(hook, whPersister, nil)
	err := dbHook.Reset()
	s.NoError(err)

	updatedHook, err := whPersister.Get(hook.ID)
	s.Require().NoError(err)

	s.Less(updatedHook.Failures, hook.Failures, "Failures should be reset to 0")
	s.Equal(0, updatedHook.Failures)

	s.True(updatedHook.ExpiresAt.After(now))
	s.True(updatedHook.UpdatedAt.After(now))
}

func (s *databaseHookSuite) TestDatabaseHook_IsEnabled() {
	hook, whPersister := s.loadWebhook("a47fe92a-1e4b-4119-8653-55ad82737c88")

	dbHook := NewDatabaseHook(hook, whPersister, nil)

	s.True(dbHook.IsEnabled())
}

func (s *databaseHookSuite) TestDatabaseHook_IsDisabled() {
	hook, whPersister := s.loadWebhook("279beae1-8a6d-4eaf-a791-1fa79d21d37a")

	dbHook := NewDatabaseHook(hook, whPersister, nil)

	s.False(dbHook.IsEnabled())
}

func (s *databaseHookSuite) loadWebhook(hookId string) (models.Webhook, persistence.WebhookPersister) {
	err := s.LoadFixtures("../test/fixtures/webhooks")
	s.Require().NoError(err)

	whPersister := s.Storage.GetWebhookPersister(nil)
	hook, err := whPersister.Get(uuid.FromStringOrNil(hookId))
	s.Require().NoError(err)
	s.Require().NotEmpty(hook)

	return *hook, whPersister
}
