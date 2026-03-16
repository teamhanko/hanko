package validation

import (
	"fmt"
	"net"
)

// SecurityMode defines the outbound destination policy for webhook callbacks.
type SecurityMode string

const (
	SecurityModePublicOnly SecurityMode = "public_only"
	SecurityModeCustom     SecurityMode = "custom"
	SecurityModeInsecure   SecurityMode = "insecure"
)

// WebhookSecurityPolicy contains the security settings for webhook validation.
// This mirrors config.WebhookSecurity but is defined here to avoid circular dependencies.
type WebhookSecurityPolicy struct {
	Mode                  SecurityMode
	AllowedSchemes        []string
	FollowRedirects       bool
	MaxRedirects          int
	AllowedHosts          []string
	AllowedDomains        []string
	AllowedCIDRs          []string
	BlockedHosts          []string
	BlockedDomains        []string
	BlockedCIDRs          []string
	DenyMetadataEndpoints bool
}

// Validator provides shared validation logic for webhook callback URLs.
// It validates based on security policies defined in WebhookSecurityPolicy configuration.
type Validator struct {
	security WebhookSecurityPolicy
}

// NewValidator creates a new Validator instance with the given security configuration.
func NewValidator(security WebhookSecurityPolicy) *Validator {
	return &Validator{
		security: security,
	}
}

// ValidateHost validates a hostname against the security policy.
// This includes checking blocked lists and allowed lists (for custom mode).
func (v *Validator) ValidateHost(host string) error {
	host = NormalizeHost(host)
	if host == "" {
		return fmt.Errorf("host must not be empty")
	}

	// Check if host is a known metadata endpoint
	if v.security.DenyMetadataEndpoints && IsMetadataHost(host) {
		return fmt.Errorf("metadata endpoint host '%s' is blocked", host)
	}

	// Check blocked lists (applies to all modes)
	if MatchesHost(host, v.security.BlockedHosts) {
		return fmt.Errorf("host '%s' is blocked", host)
	}

	if MatchesDomain(host, v.security.BlockedDomains) {
		return fmt.Errorf("host '%s' matches a blocked domain", host)
	}

	// Check allowed lists (only in custom mode with explicit allowlists)
	if err := v.validateAllowedHostPolicy(host); err != nil {
		return err
	}

	return nil
}

// ValidateIP validates an IP address against the security policy.
// This includes checking absolute denies (metadata IPs, blocked CIDRs)
// and mode-specific rules (public-only, custom allowlists).
func (v *Validator) ValidateIP(ip net.IP) error {
	// Absolute denies apply to all modes
	if err := v.validateAbsoluteDenies(ip); err != nil {
		return err
	}

	// Mode-specific validation
	return v.validateModeDecision(ip)
}

// validateAllowedHostPolicy checks if the host is allowed based on the security mode
// and configured allowlists.
func (v *Validator) validateAllowedHostPolicy(host string) error {
	switch v.security.Mode {
	case SecurityModeInsecure:
		return nil
	case SecurityModePublicOnly:
		return nil
	case SecurityModeCustom:
		// If no allowlists are configured, allow all (except blocked)
		if len(v.security.AllowedHosts) == 0 && len(v.security.AllowedDomains) == 0 {
			return nil
		}

		// If allowlists exist, host must match
		if MatchesHost(host, v.security.AllowedHosts) || MatchesDomain(host, v.security.AllowedDomains) {
			return nil
		}

		return fmt.Errorf("host '%s' is not in the allowed host/domain list", host)
	default:
		return fmt.Errorf("unsupported webhook security mode '%s'", v.security.Mode)
	}
}

// validateAbsoluteDenies checks for conditions that block the IP regardless of mode.
func (v *Validator) validateAbsoluteDenies(ip net.IP) error {
	if IPMatchesCIDRs(ip, v.security.BlockedCIDRs) {
		return fmt.Errorf("IP '%s' matches a blocked CIDR", ip.String())
	}

	if v.security.DenyMetadataEndpoints && IsMetadataIP(ip) {
		return fmt.Errorf("metadata endpoint IP '%s' is blocked", ip.String())
	}

	return nil
}

// validateModeDecision validates the IP based on the configured security mode.
func (v *Validator) validateModeDecision(ip net.IP) error {
	switch v.security.Mode {
	case SecurityModeInsecure:
		return nil
	case SecurityModePublicOnly:
		return v.validatePublicOnly(ip)
	case SecurityModeCustom:
		return v.validateCustom(ip)
	default:
		return fmt.Errorf("unsupported webhook security mode '%s'", v.security.Mode)
	}
}

// validatePublicOnly ensures the IP is publicly routable (public_only mode).
func (v *Validator) validatePublicOnly(ip net.IP) error {
	if !IsPublicRoutableIP(ip) {
		return fmt.Errorf("non-public IP '%s' is not allowed in public_only mode", ip.String())
	}

	return nil
}

// validateCustom validates the IP in custom mode.
// Private IPs are only allowed if explicitly listed in AllowedCIDRs.
// Public IPs are allowed unless blocked.
func (v *Validator) validateCustom(ip net.IP) error {
	// If IP is in allowed CIDRs, it's explicitly permitted
	if IPMatchesCIDRs(ip, v.security.AllowedCIDRs) {
		return nil
	}

	// Non-public IPs must be explicitly allowlisted
	if !IsPublicRoutableIP(ip) {
		return fmt.Errorf("non-public IP '%s' is not allowed unless explicitly allowlisted", ip.String())
	}

	return nil
}
