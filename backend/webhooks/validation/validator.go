package validation

import (
	"fmt"
	"net"
)

// SecurityMode defines the outbound destination policy for webhook callbacks.
type SecurityMode string

const (
	SecurityModePublicOnly   SecurityMode = "public_only"
	SecurityModeInternalOnly SecurityMode = "internal_only"
	SecurityModeCustom       SecurityMode = "custom"
	SecurityModeInsecure     SecurityMode = "insecure"
)

// WebhookSecurityPolicy contains the security settings for webhook validation.
// This mirrors config.WebhookSecurity but is defined here to avoid circular dependencies.
type WebhookSecurityPolicy struct {
	Mode                     SecurityMode
	AllowedSchemes           []string
	FollowRedirects          bool
	MaxRedirects             int
	SkipResolvedIPValidation bool
	AllowedHosts             []string
	AllowedDomains           []string
	AllowedCIDRs             []string
	BlockedHosts             []string
	BlockedDomains           []string
	BlockedCIDRs             []string
	DenyMetadataEndpoints    bool
	SanitizeErrors           bool
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
		err := fmt.Errorf("host must not be empty")
		return SanitizeError(err, v.security.SanitizeErrors)
	}

	// Check if host is a known metadata endpoint
	if v.security.DenyMetadataEndpoints && IsMetadataHost(host) {
		err := fmt.Errorf("metadata endpoint host '%s' is blocked", host)
		return SanitizeError(err, v.security.SanitizeErrors)
	}

	// Check blocked lists (applies to all modes)
	if MatchesHost(host, v.security.BlockedHosts) {
		err := fmt.Errorf("host '%s' is blocked", host)
		return SanitizeError(err, v.security.SanitizeErrors)
	}

	if MatchesDomain(host, v.security.BlockedDomains) {
		err := fmt.Errorf("host '%s' matches a blocked domain", host)
		return SanitizeError(err, v.security.SanitizeErrors)
	}

	// Check allowed lists (only in custom mode with explicit allowlists)
	if err := v.validateAllowedHostPolicy(host); err != nil {
		return SanitizeError(err, v.security.SanitizeErrors)
	}

	return nil
}

// ValidateIP validates an IP address against the security policy.
// This includes checking absolute denies (metadata IPs, blocked CIDRs)
// and mode-specific rules (public-only, custom allowlists).
//
// The wasHostnameValidated parameter indicates whether this IP was resolved from
// a hostname that already passed hostname validation. When true and
// SkipResolvedIPValidation is enabled, mode-specific IP validation is skipped
// (absolute denies like blocked CIDRs still apply).
func (v *Validator) ValidateIP(ip net.IP, wasHostnameValidated bool) error {
	// Absolute denies apply to all modes
	if err := v.validateAbsoluteDenies(ip); err != nil {
		return SanitizeError(err, v.security.SanitizeErrors)
	}

	// Skip mode-specific validation if this IP was resolved from a validated hostname
	// and SkipResolvedIPValidation is enabled
	if wasHostnameValidated && v.security.SkipResolvedIPValidation {
		return nil
	}

	// Mode-specific validation
	err := v.validateModeDecision(ip)
	return SanitizeError(err, v.security.SanitizeErrors)
}

// validateAllowedHostPolicy checks if the host is allowed based on the security mode
// and configured allowlists.
func (v *Validator) validateAllowedHostPolicy(host string) error {
	switch v.security.Mode {
	case SecurityModeInsecure:
		return nil
	case SecurityModePublicOnly:
		return nil
	case SecurityModeInternalOnly:
		return nil
	case SecurityModeCustom:
		// In custom mode, if no host/domain allowlists configured, skip host validation
		// (IP validation will handle CIDR allowlists if configured)
		if len(v.security.AllowedHosts) == 0 && len(v.security.AllowedDomains) == 0 {
			return nil
		}

		// Check if host is a literal IP and matches allowed_cidrs
		// This allows literal IPs in CIDR ranges to pass even when allowed_hosts is configured
		if ip := net.ParseIP(host); ip != nil && len(v.security.AllowedCIDRs) > 0 {
			if IPMatchesCIDRs(ip, v.security.AllowedCIDRs) {
				return nil
			}
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
	case SecurityModeInternalOnly:
		return v.validateInternalOnly(ip)
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

// validateInternalOnly ensures the IP is non-public/internal (internal_only mode).
func (v *Validator) validateInternalOnly(ip net.IP) error {
	if IsPublicRoutableIP(ip) {
		return fmt.Errorf("public IP '%s' is not allowed in internal_only mode", ip.String())
	}

	return nil
}

// validateCustom validates the IP in custom mode.
// Custom mode requires explicit allowlist configuration (enforced at config validation).
// IPs can be allowed via allowed_cidrs (CIDR notation) or allowed_hosts (literal IPs).
func (v *Validator) validateCustom(ip net.IP) error {
	// Check if IP matches allowed CIDRs
	if len(v.security.AllowedCIDRs) > 0 {
		if IPMatchesCIDRs(ip, v.security.AllowedCIDRs) {
			return nil
		}
	}

	// Check if IP matches allowed hosts (for literal IPs in allowed_hosts)
	if len(v.security.AllowedHosts) > 0 {
		ipStr := ip.String()
		if MatchesHost(ipStr, v.security.AllowedHosts) {
			return nil
		}
	}

	// If no allow lists configured, this should never happen (caught by config validation)
	if len(v.security.AllowedCIDRs) == 0 && len(v.security.AllowedHosts) == 0 && len(v.security.AllowedDomains) == 0 {
		return fmt.Errorf("custom mode requires allowlist configuration")
	}

	return fmt.Errorf("IP '%s' is not in the allowed CIDR or host list", ip.String())
}
