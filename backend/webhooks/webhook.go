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
	"github.com/teamhanko/hanko/backend/v2/webhooks/validation"
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

	// Validate the callback URL and get the validated IPs to prevent DNS rebinding
	validationResult, err := validator.ValidateAndGetIPs(validateCtx, bh.Callback)
	if err != nil {
		// Log detailed error for debugging
		detailedErr := validation.GetDetailedError(err)
		bh.logError(fmt.Errorf("webhook callback rejected by outbound policy: %w", detailedErr))
		// Return sanitized error to caller
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

	// Create a custom dialer that only connects to the validated IPs
	// This prevents DNS rebinding attacks by bypassing DNS resolution in http.Client
	dialer := NewValidatedDialer(validationResult.ValidatedIPs, validationResult.Host)

	client := &http.Client{
		Timeout: requestTimeout,
		Transport: &http.Transport{
			DialContext:           dialer.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          1,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !bh.Security.FollowRedirects {
				return http.ErrUseLastResponse
			}

			if bh.Security.MaxRedirects >= 0 && len(via) > bh.Security.MaxRedirects {
				return fmt.Errorf("too many redirects")
			}

			// Validate the redirect target and get its validated IPs
			redirectCtx, redirectCancel := context.WithTimeout(req.Context(), resolveTimeout)
			defer redirectCancel()

			redirectResult, err := validator.ValidateAndGetIPs(redirectCtx, req.URL.String())
			if err != nil {
				// Log detailed error for debugging
				detailedErr := validation.GetDetailedError(err)
				bh.logError(fmt.Errorf("redirect target rejected by outbound policy: %w", detailedErr))
				// Return error with context (sanitization already applied by validator)
				return fmt.Errorf("redirect target rejected by outbound policy: %w", err)
			}

			// Update the dialer to use the new validated IPs for the redirect
			// Note: This is safe because the transport is not shared across goroutines
			dialer.validatedIPs = redirectResult.ValidatedIPs
			dialer.originalHost = redirectResult.Host
			dialer.currentIPIndex = 0

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
