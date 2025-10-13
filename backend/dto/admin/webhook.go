package admin

import (
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
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
