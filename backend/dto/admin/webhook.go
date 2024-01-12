package admin

import (
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/webhooks/events"
)

type WebhookListResponseDto struct {
	Database models.Webhooks `json:"database"`
	Config   config.Webhooks `json:"config"`
}

type CreateWebhookRequestDto struct {
	Callback string        `json:"callback" validate:"required,url"`
	Events   events.Events `json:"events" validate:"required,min=1,dive,hanko_event"`
}

type GetWebhookRequestDto struct {
	ID string `param:"id" validate:"required,uuid4"`
}

type UpdateWebhookRequestDto struct {
	GetWebhookRequestDto
	CreateWebhookRequestDto
	Enabled bool `json:"enabled" validate:"required,boolean"`
}
