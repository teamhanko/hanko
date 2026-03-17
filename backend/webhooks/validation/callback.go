package validation

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
)

// WebhookSecurityConfig is an interface that provides access to webhook security settings.
// This allows us to accept config.WebhookSecurity without importing the config package.
type WebhookSecurityConfig interface {
	ToWebhookSecurityPolicy() WebhookSecurityPolicy
	GetAllowedSchemes() []string
}

// ValidateWebhook validates a webhook's callback URL and associated events
// against the provided security policy.
// This is the main entry point for webhook validation that can be used by both
// config webhooks and database webhooks.
func ValidateWebhook(callback string, eventsStr []string, security WebhookSecurityConfig) error {
	parsed, err := url.Parse(callback)
	if err != nil {
		return fmt.Errorf("callback is not a valid URL: %w", err)
	}

	if err := validateParsedURL(parsed, security); err != nil {
		return err
	}

	if err := validateLiteralIP(parsed, security); err != nil {
		return err
	}

	if err := validateEvents(eventsStr); err != nil {
		return err
	}

	return nil
}

// validateParsedURL validates the structure and security of a parsed URL.
func validateParsedURL(parsed *url.URL, security WebhookSecurityConfig) error {
	if parsed.Scheme == "" {
		return fmt.Errorf("callback URL must include a scheme")
	}

	if parsed.Host == "" {
		return fmt.Errorf("callback URL must include a host")
	}

	if parsed.User != nil {
		return fmt.Errorf("callback URL must not include user info")
	}

	allowedSchemes := security.GetAllowedSchemes()
	schemeAllowed := false
	for _, scheme := range allowedSchemes {
		if strings.EqualFold(strings.TrimSpace(scheme), parsed.Scheme) {
			schemeAllowed = true
			break
		}
	}

	if !schemeAllowed {
		return fmt.Errorf("callback scheme '%s' is not allowed", parsed.Scheme)
	}

	validator := NewValidator(security.ToWebhookSecurityPolicy())
	host := parsed.Hostname()

	if err := validator.ValidateHost(host); err != nil {
		return fmt.Errorf("callback %w", err)
	}

	return nil
}

// validateLiteralIP validates that a callback URL using a literal IP address is allowed.
func validateLiteralIP(parsed *url.URL, security WebhookSecurityConfig) error {
	host := NormalizeHost(parsed.Hostname())
	ip := net.ParseIP(host)
	if ip == nil {
		return nil
	}

	validator := NewValidator(security.ToWebhookSecurityPolicy())
	if err := validator.ValidateIP(ip); err != nil {
		return fmt.Errorf("callback %w", err)
	}

	return nil
}

// validateEvents validates that all provided event names are valid webhook events.
func validateEvents(eventsStr []string) error {
	if len(eventsStr) > 0 {
		for i, e := range eventsStr {
			isValid := events.IsValidEvent(events.Event(e))
			if !isValid {
				return fmt.Errorf("event '%s' at events[%d] is not a valid webhook event", e, i)
			}
		}
	}

	return nil
}
