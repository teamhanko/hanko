package webhooks

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"time"
)

const (
	FailureExpireRate = 5
)

type DatabaseHook struct {
	BaseWebhook
	persister persistence.WebhookPersister
	rawHook   models.Webhook
}

func NewDatabaseHook(dbHook models.Webhook, persister persistence.WebhookPersister, logger echo.Logger) Webhook {
	return &DatabaseHook{
		BaseWebhook{
			Logger:   logger,
			Callback: dbHook.Callback,
			Events:   events.ConvertFromDbList(dbHook.WebhookEvents),
		},
		persister,
		dbHook,
	}
}

func (dh *DatabaseHook) DisableOnExpiryDate(now time.Time) error {
	// check expire date
	if dh.rawHook.ExpiresAt.Before(now) {
		dh.rawHook.Enabled = false

		err := dh.persister.Update(dh.rawHook)
		if err != nil {
			dh.Logger.Error(fmt.Errorf("unable to expire webhook on date: %w", err))
			return err
		}
	}

	return nil
}

func (dh *DatabaseHook) DisableOnFailure() error {
	// increase Failure-Rate
	dh.rawHook.Failures++

	if dh.rawHook.Failures > FailureExpireRate {
		dh.rawHook.Enabled = false
	}

	err := dh.persister.Update(dh.rawHook)
	if err != nil {
		dh.Logger.Error(fmt.Errorf("unable to expire webhook on failure: %w", err))
		return err
	}

	return nil
}

func (dh *DatabaseHook) Reset() error {
	now := time.Now()
	dh.rawHook.Failures = 0
	dh.rawHook.ExpiresAt = now.Add(WebhookExpireDuration)
	dh.rawHook.UpdatedAt = now

	err := dh.persister.Update(dh.rawHook)
	if err != nil {
		dh.Logger.Error(fmt.Errorf("unable to reset webhook failures: %w", err))
		return err
	}

	return nil
}

func (dh *DatabaseHook) IsEnabled() bool {
	return dh.rawHook.Enabled
}
