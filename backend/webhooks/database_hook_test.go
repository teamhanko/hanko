package webhooks

import (
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"testing"
	"time"
)

func TestNewDatabaseHook(t *testing.T) {
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

	hookId, err := uuid.NewV4()
	assert.NoError(t, err)

	hook := models.Webhook{
		ID:        hookId,
		Enabled:   false,
		Failures:  0,
		ExpiresAt: time.Now().Add(24 * -1 * time.Hour),
	}

	dbHook := NewDatabaseHook(hook, persister.GetWebhookPersister(nil), nil)
	require.NotEmpty(t, dbHook)
}

func TestDatabaseHook_DisableOnExpiryDate(t *testing.T) {
	hookId, err := uuid.NewV4()
	assert.NoError(t, err)

	now := time.Now()

	hook := models.Webhook{
		ID:        hookId,
		Enabled:   true,
		Failures:  0,
		ExpiresAt: now.Add(24 * -1 * time.Hour),
	}

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
		models.Webhooks{hook},
		nil,
	)

	whPersister := persister.GetWebhookPersister(nil)

	dbHook := NewDatabaseHook(hook, whPersister, nil)
	err = dbHook.DisableOnExpiryDate(now)
	assert.NoError(t, err)

	updatedHook, err := whPersister.Get(hook.ID)
	assert.NoError(t, err)

	require.False(t, updatedHook.Enabled)
}
func TestDatabaseHook_DoNotDisableOnExpiryDate(t *testing.T) {
	hookId, err := uuid.NewV4()
	assert.NoError(t, err)

	now := time.Now()

	hook := models.Webhook{
		ID:        hookId,
		Enabled:   true,
		Failures:  0,
		ExpiresAt: now.Add(24 * 1 * time.Hour),
	}

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
		models.Webhooks{hook},
		nil,
	)

	whPersister := persister.GetWebhookPersister(nil)

	dbHook := NewDatabaseHook(hook, whPersister, nil)
	err = dbHook.DisableOnExpiryDate(now)
	assert.NoError(t, err)

	updatedHook, err := whPersister.Get(hook.ID)
	assert.NoError(t, err)

	require.True(t, updatedHook.Enabled)
}

func TestDatabaseHook_DisableOnFailure(t *testing.T) {
	hookId, err := uuid.NewV4()
	assert.NoError(t, err)

	hook := models.Webhook{
		ID:       hookId,
		Enabled:  true,
		Failures: FailureExpireRate,
	}

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
		models.Webhooks{hook},
		nil,
	)

	whPersister := persister.GetWebhookPersister(nil)

	dbHook := NewDatabaseHook(hook, whPersister, nil)
	err = dbHook.DisableOnFailure()
	assert.NoError(t, err)

	updatedHook, err := whPersister.Get(hook.ID)
	assert.NoError(t, err)

	require.False(t, updatedHook.Enabled)
}

func TestDatabaseHook_DoNotDisableOnFailure(t *testing.T) {
	hookId, err := uuid.NewV4()
	assert.NoError(t, err)

	hook := models.Webhook{
		ID:       hookId,
		Enabled:  true,
		Failures: 0,
	}

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
		models.Webhooks{hook},
		nil,
	)

	whPersister := persister.GetWebhookPersister(nil)

	dbHook := NewDatabaseHook(hook, whPersister, nil)
	err = dbHook.DisableOnFailure()
	assert.NoError(t, err)

	updatedHook, err := whPersister.Get(hook.ID)
	assert.NoError(t, err)

	require.True(t, updatedHook.Enabled)
}

func TestDatabaseHook_Reset(t *testing.T) {
	hookId, err := uuid.NewV4()
	assert.NoError(t, err)

	now := time.Now()

	hook := models.Webhook{
		ID:       hookId,
		Enabled:  true,
		Failures: 3,
	}

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
		models.Webhooks{hook},
		nil,
	)

	whPersister := persister.GetWebhookPersister(nil)

	dbHook := NewDatabaseHook(hook, whPersister, nil)
	err = dbHook.Reset()
	assert.NoError(t, err)

	updatedHook, err := whPersister.Get(hook.ID)
	assert.NoError(t, err)

	require.Less(t, updatedHook.Failures, hook.Failures, "Failures should be reset to 0")
	require.Equal(t, 0, updatedHook.Failures)

	require.True(t, updatedHook.ExpiresAt.After(now))
	require.True(t, updatedHook.UpdatedAt.After(now))
}

func TestDatabaseHook_IsEnabled(t *testing.T) {
	hookId, err := uuid.NewV4()
	assert.NoError(t, err)

	hook := models.Webhook{
		ID:      hookId,
		Enabled: true,
	}

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
		models.Webhooks{hook},
		nil,
	)

	whPersister := persister.GetWebhookPersister(nil)

	dbHook := NewDatabaseHook(hook, whPersister, nil)

	require.True(t, dbHook.IsEnabled())
}

func TestDatabaseHook_IsDisabled(t *testing.T) {
	hookId, err := uuid.NewV4()
	assert.NoError(t, err)

	hook := models.Webhook{
		ID:      hookId,
		Enabled: false,
	}

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
		models.Webhooks{hook},
		nil,
	)

	whPersister := persister.GetWebhookPersister(nil)

	dbHook := NewDatabaseHook(hook, whPersister, nil)

	require.False(t, dbHook.IsEnabled())
}
