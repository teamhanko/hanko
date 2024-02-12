package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"

	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type WebhookPersister interface {
	Create(webhook models.Webhook, events models.WebhookEvents) error
	Update(webhook models.Webhook) error
	Delete(webhook models.Webhook) error
	AddEvent(event models.WebhookEvent) error
	RemoveEvent(event models.WebhookEvent) error
	List(includeDisabled bool) (models.Webhooks, error)
	Get(webhookId uuid.UUID) (*models.Webhook, error)
}

type webhookPersister struct {
	db *pop.Connection
}

func NewWebhookPersister(db *pop.Connection) WebhookPersister {
	return &webhookPersister{db: db}
}

func (w *webhookPersister) Create(webhook models.Webhook, events models.WebhookEvents) error {
	vErr, err := w.db.ValidateAndCreate(&webhook)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("webhook object validation failed: %w", vErr)
	}

	for _, event := range events {
		vErr, err = w.db.ValidateAndCreate(&event)
		if err != nil {
			return fmt.Errorf("failed to create webhook event: %w", err)
		}

		if vErr != nil && vErr.HasAny() {
			return fmt.Errorf("webhook event object validation failed: %w", vErr)
		}
	}

	return nil
}

func (w *webhookPersister) Update(webhook models.Webhook) error {
	vErr, err := w.db.ValidateAndUpdate(&webhook)
	if err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("webhook object validation failed: %w", vErr)
	}

	return nil
}

func (w *webhookPersister) Delete(webhook models.Webhook) error {
	err := w.db.Destroy(&webhook)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	return nil
}

func (w *webhookPersister) AddEvent(event models.WebhookEvent) error {
	vErr, err := w.db.ValidateAndCreate(&event)
	if err != nil {
		return fmt.Errorf("failed to create webhook event: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("webhook event object validation failed: %w", vErr)
	}

	return nil
}

func (w *webhookPersister) RemoveEvent(event models.WebhookEvent) error {
	err := w.db.Destroy(&event)
	if err != nil {
		return fmt.Errorf("failed to remove webhook event: %w", err)
	}

	return nil
}

func (w *webhookPersister) List(includeDisabled bool) (models.Webhooks, error) {
	webhooks := make(models.Webhooks, 0)
	query := w.db.Eager().Q()
	if !includeDisabled {
		query = query.Where("enabled = true")
	}
	err := query.All(&webhooks)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return webhooks, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch webhooks: %w", err)
	}

	return webhooks, nil
}

func (w *webhookPersister) Get(webhookId uuid.UUID) (*models.Webhook, error) {
	webhook := models.Webhook{}
	err := w.db.Eager().Find(&webhook, webhookId)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	return &webhook, nil
}
