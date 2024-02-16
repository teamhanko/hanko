package test

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewWebhookPersister(initHooks models.Webhooks, initEvents models.WebhookEvents) persistence.WebhookPersister {
	return &TestWebhookPersister{
		append(models.Webhooks{}, initHooks...),
		append(models.WebhookEvents{}, initEvents...),
	}
}

type TestWebhookPersister struct {
	webhooks models.Webhooks
	events   models.WebhookEvents
}

func (w *TestWebhookPersister) Create(webhook models.Webhook, events models.WebhookEvents) error {
	webhook.WebhookEvents = events
	w.webhooks = append(w.webhooks, webhook)
	w.events = append(w.events, events...)

	return nil
}

func (w *TestWebhookPersister) Update(webhook models.Webhook) error {
	for i, hook := range w.webhooks {
		if hook.ID == webhook.ID {
			w.webhooks[i] = webhook
		}
	}

	return nil
}

func (w *TestWebhookPersister) Delete(webhook models.Webhook) error {
	index := -1
	for i, hook := range w.webhooks {
		if hook.ID == webhook.ID {
			index = i
		}
	}
	if index > -1 {
		w.webhooks = append(w.webhooks[:index], w.webhooks[index+1:]...)
	}

	return nil
}

func (w *TestWebhookPersister) AddEvent(event models.WebhookEvent) error {
	w.events = append(w.events, event)

	return nil
}

func (w *TestWebhookPersister) RemoveEvent(event models.WebhookEvent) error {
	index := -1

	for i, commonEvent := range w.events {
		if commonEvent.ID == event.ID {
			index = i
		}
	}

	for _, hook := range w.webhooks {
		hookIndex := -1
		for i, hookEvent := range hook.WebhookEvents {
			if hookEvent.ID == w.events[index].ID {
				hookIndex = i
			}
		}

		if hookIndex > -1 {
			w.events = append(w.events[:hookIndex], w.events[hookIndex+1:]...)
		}
	}

	if index > -1 {
		w.events = append(w.events[:index], w.events[index+1:]...)
	}

	return nil
}

func (w *TestWebhookPersister) List(includeDisabled bool) (models.Webhooks, error) {
	list := make(models.Webhooks, 0)
	for _, hook := range w.webhooks {
		if !includeDisabled && hook.Enabled == false {
			continue
		}

		list = append(list, hook)
	}

	return list, nil
}

func (w *TestWebhookPersister) Get(webhookId uuid.UUID) (*models.Webhook, error) {
	for _, hook := range w.webhooks {
		if hook.ID == webhookId {
			return &hook, nil
		}
	}

	return nil, nil
}
