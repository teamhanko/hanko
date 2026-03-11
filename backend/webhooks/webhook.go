package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
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
	Logger         echo.Logger
	Callback       string
	Events         events.Events
	Security       config.WebhookSecurity
	RequestTimeout time.Duration
	ResolveTimeout time.Duration
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
	validator := NewURLPolicyValidator(bh.Security)

	resolveTimeout := bh.ResolveTimeout
	if resolveTimeout == 0 {
		resolveTimeout = 5 * time.Second
	}

	validateCtx, cancel := context.WithTimeout(context.Background(), resolveTimeout)
	defer cancel()

	if err := validator.Validate(validateCtx, bh.Callback); err != nil {
		bh.logError(fmt.Errorf("webhook callback rejected by outbound policy: %w", err))
		return err
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		bh.logError(fmt.Errorf("unable to marshal webhook payload: %w", err))
		return err
	}

	req, err := http.NewRequest(http.MethodPost, bh.Callback, bytes.NewReader(dataJSON))
	if err != nil {
		bh.logError(fmt.Errorf("unable to create webhook request: %w", err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	requestTimeout := bh.RequestTimeout
	if requestTimeout == 0 {
		requestTimeout = 10 * time.Second
	}

	client := &http.Client{
		Timeout: requestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !bh.Security.FollowRedirects {
				return http.ErrUseLastResponse
			}

			if bh.Security.MaxRedirects >= 0 && len(via) > bh.Security.MaxRedirects {
				return fmt.Errorf("too many redirects")
			}

			redirectCtx, redirectCancel := context.WithTimeout(req.Context(), resolveTimeout)
			defer redirectCancel()

			if err := validator.Validate(redirectCtx, req.URL.String()); err != nil {
				return fmt.Errorf("redirect target rejected by outbound policy: %w", err)
			}

			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		bh.logError(fmt.Errorf("unable to execute webhook request: %w", err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusMultipleChoices {
		err := fmt.Errorf("webhook request failed due to status code: %d", resp.StatusCode)
		bh.logError(err)
		return err
	}

	return nil
}

func (bh *BaseWebhook) logError(err error) {
	if bh.Logger != nil {
		bh.Logger.Error(err)
	}
}
