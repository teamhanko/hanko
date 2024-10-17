package handler

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto/admin"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/webhooks"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"net/http"
	"time"
)

type WebhookHandler interface {
	List(ctx echo.Context) error
	Create(ctx echo.Context) error
	Get(ctx echo.Context) error
	Delete(ctx echo.Context) error
	Update(ctx echo.Context) error
}

const (
	uuidErrorFormat = "unable to create uuid: %w"
)

type webhookHandler struct {
	cfg       config.WebhookSettings
	persister persistence.Persister
}

func NewWebhookHandler(cfg config.WebhookSettings, persister persistence.Persister) WebhookHandler {
	return &webhookHandler{
		cfg:       cfg,
		persister: persister,
	}
}

func (w *webhookHandler) List(ctx echo.Context) error {
	persister := w.persister.GetWebhookPersister(nil)
	dbHooks, err := persister.List(true)
	if err != nil {
		ctx.Logger().Error(err)
		return fmt.Errorf("failed to list users: %w", err)
	}

	listDto := admin.WebhookListResponseDto{
		Database: dbHooks,
		Config:   w.cfg.Hooks,
	}

	return ctx.JSON(http.StatusOK, listDto)
}

func (w *webhookHandler) Create(ctx echo.Context) error {
	var dto admin.CreateWebhookRequestDto
	err := ctx.Bind(&dto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = ctx.Validate(dto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	now := time.Now()

	newUuid, err := uuid.NewV4()
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf(uuidErrorFormat, err))
	}

	model := models.Webhook{
		ID:            newUuid,
		Callback:      dto.Callback,
		Enabled:       true,
		Failures:      0,
		ExpiresAt:     now.Add(webhooks.WebhookExpireDuration), // 30 Days from now
		WebhookEvents: nil,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	dbEvents, err := w.createWebhookEvents(dto.Events, model, now)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf(uuidErrorFormat, err))
	}

	persister := w.persister.GetWebhookPersister(nil)
	err = persister.Create(model, dbEvents)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("unable to save webhook: %w", err))
	}

	model.WebhookEvents = dbEvents

	return ctx.JSON(http.StatusCreated, model)
}

func (w *webhookHandler) createWebhookEvents(evts events.Events, webhook models.Webhook, now time.Time) (models.WebhookEvents, error) {
	eventList := make(models.WebhookEvents, 0)
	for _, event := range evts {
		newUuid, err := uuid.NewV4()
		if err != nil {
			return eventList, err
		}

		model := models.WebhookEvent{
			ID:        newUuid,
			Webhook:   &webhook,
			Event:     string(event),
			CreatedAt: now,
			UpdatedAt: now,
		}

		eventList = append(eventList, model)
	}

	return eventList, nil
}

func (w *webhookHandler) Get(ctx echo.Context) error {
	var dto admin.GetWebhookRequestDto
	err := ctx.Bind(&dto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = ctx.Validate(dto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	webhookId, _ := uuid.FromString(dto.ID)
	webhook, err := w.getWebhook(webhookId, w.persister.GetWebhookPersister(nil))
	if err != nil {
		ctx.Logger().Error(err)
		return err
	}

	return ctx.JSON(http.StatusOK, webhook)
}

func (w *webhookHandler) Delete(ctx echo.Context) error {
	var dto admin.GetWebhookRequestDto
	err := ctx.Bind(&dto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = ctx.Validate(dto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	persister := w.persister.GetWebhookPersister(nil)

	webhookId, _ := uuid.FromString(dto.ID)
	webhook, err := w.getWebhook(webhookId, persister)
	if err != nil {
		ctx.Logger().Error(err)
		return err
	}

	err = persister.Delete(*webhook)
	if err != nil {
		ctx.Logger().Error(err)
		return fmt.Errorf("unable to delete webhook from database: %w", err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (w *webhookHandler) Update(ctx echo.Context) error {
	var dto admin.UpdateWebhookRequestDto
	err := ctx.Bind(&dto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = ctx.Validate(dto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	return w.persister.Transaction(func(tx *pop.Connection) error {
		persister := w.persister.GetWebhookPersister(tx)

		webhookId, _ := uuid.FromString(dto.ID)

		webhook, err := w.getWebhook(webhookId, persister)
		if err != nil {
			ctx.Logger().Error(err)
			return err
		}

		for _, event := range webhook.WebhookEvents {
			err := persister.RemoveEvent(event)
			if err != nil {
				ctx.Logger().Error(err)
				return fmt.Errorf("unable to delete event: %w", err)
			}
		}

		now := time.Now()
		dbEvents, err := w.createWebhookEvents(dto.Events, *webhook, now)
		if err != nil {
			ctx.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf(uuidErrorFormat, err))
		}

		webhook.WebhookEvents = dbEvents
		webhook.Callback = dto.Callback
		webhook.UpdatedAt = now
		webhook.Enabled = dto.Enabled
		webhook.Failures = 0
		webhook.ExpiresAt = now.Add(webhooks.WebhookExpireDuration)

		err = persister.Update(*webhook)
		if err != nil {
			ctx.Logger().Error(err)
			return fmt.Errorf("unable to update webhook: %w", err)
		}

		return ctx.JSON(http.StatusOK, webhook)
	})
}

func (w *webhookHandler) getWebhook(id uuid.UUID, persister persistence.WebhookPersister) (*models.Webhook, error) {
	webhook, err := persister.Get(id)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch webhook from database: %w", err)
	}

	if webhook == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "unable to find webhook with id: %s", id.String())
	}

	return webhook, nil
}
