package webhooks

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/webhooks/validation"
)

type URLPolicyValidator struct {
	security config.WebhookSecurity
	resolver *net.Resolver
}

func NewURLPolicyValidator(security config.WebhookSecurity) *URLPolicyValidator {
	return &URLPolicyValidator{
		security: security,
		resolver: net.DefaultResolver,
	}
}

// ValidationResult contains the result of URL validation including validated IPs
// that can be used to prevent DNS rebinding attacks.
type ValidationResult struct {
	// ValidatedIPs contains the IP addresses that passed security validation
	ValidatedIPs []net.IP
	// Host is the normalized hostname from the URL
	Host string
}

// ValidateAndGetIPs validates the URL and returns the validated IP addresses.
// This allows callers to pin connections to validated IPs, preventing DNS rebinding.
func (v *URLPolicyValidator) ValidateAndGetIPs(ctx context.Context, rawURL string) (*ValidationResult, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook callback URL: %w", err)
	}

	host, err := v.validateParsedURL(parsed)
	if err != nil {
		return nil, err
	}

	// If the host is already a literal IP, return it directly
	if ip := net.ParseIP(host); ip != nil {
		if err := v.validateResolvedIP(ip); err != nil {
			return nil, err
		}
		return &ValidationResult{
			ValidatedIPs: []net.IP{ip},
			Host:         host,
		}, nil
	}

	// Resolve the hostname and validate all returned IPs
	ips, err := v.resolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve webhook callback host '%s': %w", host, err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("webhook callback host '%s' did not resolve to any IP addresses", host)
	}

	// All resolved IPs must satisfy the outbound policy.
	// Rejecting on any disallowed IP avoids mixed public/internal DNS answers.
	for _, ip := range ips {
		if err := v.validateResolvedIP(ip); err != nil {
			return nil, fmt.Errorf("resolved IP '%s' for host '%s' is not allowed: %w", ip.String(), host, err)
		}
	}

	return &ValidationResult{
		ValidatedIPs: ips,
		Host:         host,
	}, nil
}

// Validate validates the URL without returning the validated IPs.
// This is kept for backward compatibility but ValidateAndGetIPs is preferred
// for protection against DNS rebinding attacks.
func (v *URLPolicyValidator) Validate(ctx context.Context, rawURL string) error {
	_, err := v.ValidateAndGetIPs(ctx, rawURL)
	return err
}

func (v *URLPolicyValidator) validateParsedURL(parsed *url.URL) (string, error) {
	if parsed.Scheme == "" {
		return "", fmt.Errorf("webhook callback URL must include a scheme")
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("webhook callback URL must include a host")
	}

	if parsed.User != nil {
		return "", fmt.Errorf("webhook callback URL must not include user info")
	}

	schemeAllowed := false
	for _, scheme := range v.security.AllowedSchemes {
		if strings.EqualFold(strings.TrimSpace(scheme), parsed.Scheme) {
			schemeAllowed = true
			break
		}
	}

	if !schemeAllowed {
		return "", fmt.Errorf("webhook callback scheme '%s' is not allowed", parsed.Scheme)
	}

	host := parsed.Hostname()
	validator := validation.NewValidator(v.security.ToWebhookSecurityPolicy())

	if err := validator.ValidateHost(host); err != nil {
		return "", fmt.Errorf("webhook callback %w", err)
	}

	return validation.NormalizeHost(host), nil
}

func (v *URLPolicyValidator) validateResolvedIP(ip net.IP) error {
	validator := validation.NewValidator(v.security.ToWebhookSecurityPolicy())
	return validator.ValidateIP(ip)
}
