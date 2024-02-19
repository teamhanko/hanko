package webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"net/http"
	"strings"
	"time"
)

type Webhook interface {
	Trigger(data JobData) error
	DisableOnExpiryDate(now time.Time) error
	DisableOnFailure() error
	Reset() error
	IsEnabled() bool
	HasEvent(evt events.Event) bool
}

type Webhooks []Webhook

const (
	WebhookExpireDuration = 30 * 24 * time.Hour // 30 Days
)

type BaseWebhook struct {
	Logger   echo.Logger
	Callback string
	Events   events.Events
}

func (bh *BaseWebhook) HasEvent(evt events.Event) bool {
	for _, event := range bh.Events {
		if strings.HasPrefix(string(evt), string(event)) {
			return true
		}
	}

	return false
}

func (bh *BaseWebhook) Trigger(data JobData) error {
	// create request
	dataJson, err := json.Marshal(data)
	if err != nil {
		bh.Logger.Error(fmt.Errorf("unable to convert JobData to json: %w", err))
		return err
	}

	request, err := http.NewRequest(http.MethodPost, bh.Callback, bytes.NewReader(dataJson))
	if err != nil {
		bh.Logger.Error(fmt.Errorf("unable to create request for webhook: %w", err))
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		bh.Logger.Error(fmt.Errorf("unable to execute webhook request: %w", err))
		return err
	}

	if response.StatusCode >= http.StatusBadRequest {
		err := fmt.Errorf("request failed due to status code: %d", response.StatusCode)
		bh.Logger.Error(err)

		return err
	}

	return nil
}
