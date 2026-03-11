package webhooks

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/teamhanko/hanko/backend/v2/config"
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

func (v *URLPolicyValidator) Validate(ctx context.Context, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid webhook callback URL: %w", err)
	}

	host, err := v.validateParsedURL(parsed)
	if err != nil {
		return err
	}

	if ip := net.ParseIP(host); ip != nil {
		return v.validateResolvedIP(ip)
	}

	ips, err := v.resolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return fmt.Errorf("failed to resolve webhook callback host '%s': %w", host, err)
	}

	if len(ips) == 0 {
		return fmt.Errorf("webhook callback host '%s' did not resolve to any IP addresses", host)
	}

	// All resolved IPs must satisfy the outbound policy.
	// Rejecting on any disallowed IP avoids mixed public/internal DNS answers.
	for _, ip := range ips {
		if err := v.validateResolvedIP(ip); err != nil {
			return fmt.Errorf("resolved IP '%s' for host '%s' is not allowed: %w", ip.String(), host, err)
		}
	}

	return nil
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

	host := normalizePolicyHost(parsed.Hostname())
	if host == "" {
		return "", fmt.Errorf("webhook callback host must not be empty")
	}

	if matchesHost(host, v.security.BlockedHosts) {
		return "", fmt.Errorf("webhook callback host '%s' is blocked", host)
	}

	if matchesDomain(host, v.security.BlockedDomains) {
		return "", fmt.Errorf("webhook callback host '%s' matches a blocked domain", host)
	}

	if err := v.validateAllowedHostPolicy(host); err != nil {
		return "", err
	}

	return host, nil
}

func (v *URLPolicyValidator) validateAllowedHostPolicy(host string) error {
	switch v.security.Mode {
	case config.WebhookSecurityModeInsecure:
		return nil
	case config.WebhookSecurityModePublicOnly:
		return nil
	case config.WebhookSecurityModeCustom:
		if len(v.security.AllowedHosts) == 0 && len(v.security.AllowedDomains) == 0 {
			return nil
		}

		if matchesHost(host, v.security.AllowedHosts) || matchesDomain(host, v.security.AllowedDomains) {
			return nil
		}

		return fmt.Errorf("webhook callback host '%s' is not in the allowed host/domain list", host)
	default:
		return fmt.Errorf("unsupported webhook security mode '%s'", v.security.Mode)
	}
}

func (v *URLPolicyValidator) validateResolvedIP(ip net.IP) error {
	if err := v.validateAbsoluteDenies(ip); err != nil {
		return err
	}

	return v.validateModeDecision(ip)
}

func (v *URLPolicyValidator) validateAbsoluteDenies(ip net.IP) error {
	if ipMatchesCIDRs(ip, v.security.BlockedCIDRs) {
		return fmt.Errorf("IP '%s' matches a blocked CIDR", ip.String())
	}

	if v.security.DenyMetadataEndpoints && isMetadataIP(ip) {
		return fmt.Errorf("metadata endpoint IP '%s' is blocked", ip.String())
	}

	return nil
}

func (v *URLPolicyValidator) validateModeDecision(ip net.IP) error {
	switch v.security.Mode {
	case config.WebhookSecurityModeInsecure:
		return nil
	case config.WebhookSecurityModePublicOnly:
		return v.validatePublicOnly(ip)
	case config.WebhookSecurityModeCustom:
		return v.validateCustom(ip)
	default:
		return fmt.Errorf("unsupported webhook security mode '%s'", v.security.Mode)
	}
}

func (v *URLPolicyValidator) validatePublicOnly(ip net.IP) error {
	if !isPublicRoutableIP(ip) {
		return fmt.Errorf("non-public IP '%s' is not allowed in public_only mode", ip.String())
	}

	return nil
}

func (v *URLPolicyValidator) validateCustom(ip net.IP) error {
	if ipMatchesCIDRs(ip, v.security.AllowedCIDRs) {
		return nil
	}

	if !isPublicRoutableIP(ip) {
		return fmt.Errorf("non-public IP '%s' is not allowed unless explicitly allowlisted", ip.String())
	}

	return nil
}

func normalizePolicyHost(value string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
}

func matchesHost(host string, values []string) bool {
	host = normalizePolicyHost(host)
	for _, value := range values {
		if host == normalizePolicyHost(value) {
			return true
		}
	}
	return false
}

func matchesDomain(host string, domains []string) bool {
	host = normalizePolicyHost(host)
	for _, domain := range domains {
		normalized := normalizePolicyHost(domain)
		if host == normalized || strings.HasSuffix(host, "."+normalized) {
			return true
		}
	}
	return false
}

func ipMatchesCIDRs(ip net.IP, cidrs []string) bool {
	for _, value := range cidrs {
		_, network, err := net.ParseCIDR(strings.TrimSpace(value))
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

func isPublicRoutableIP(ip net.IP) bool {
	if ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsMulticast() ||
		ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		isReservedIP(ip) ||
		isMetadataIP(ip) {
		return false
	}

	return true
}

func isReservedIP(ip net.IP) bool {
	return ipMatchesCIDRs(ip, []string{
		"0.0.0.0/8",
		"100.64.0.0/10",
		"192.0.0.0/24",
		"192.0.2.0/24",
		"198.18.0.0/15",
		"198.51.100.0/24",
		"203.0.113.0/24",
		"240.0.0.0/4",
		"::/128",
		"100::/64",
		"2001:db8::/32",
	})
}

func isMetadataIP(ip net.IP) bool {
	return ipMatchesCIDRs(ip, []string{
		"169.254.169.254/32",
	})
}
