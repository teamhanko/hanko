package webhooks

import (
	"context"
	"net"
	"testing"

	"github.com/foxcpp/go-mockdns"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/v2/config"
)

func TestURLPolicyValidator_Validate_RejectsInvalidURL(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeInsecure,
		AllowedSchemes: []string{"http", "https"},
	})

	err := validator.Validate(context.Background(), "://broken")

	require.Error(t, err)
	require.ErrorContains(t, err, "invalid webhook callback URL")
}

func TestURLPolicyValidator_Validate_RejectsMissingScheme(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeInsecure,
		AllowedSchemes: []string{"http", "https"},
	})

	err := validator.Validate(context.Background(), "//example.com")

	require.Error(t, err)
	require.ErrorContains(t, err, "must include a scheme")
}

func TestURLPolicyValidator_Validate_RejectsDisallowedScheme(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeInsecure,
		AllowedSchemes: []string{"https"},
	})

	err := validator.Validate(context.Background(), "http://127.0.0.1/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "scheme")
}

func TestURLPolicyValidator_Validate_RejectsUserInfo(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeInsecure,
		AllowedSchemes: []string{"http", "https"},
	})

	err := validator.Validate(context.Background(), "http://user:pass@127.0.0.1/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "user info")
}

// Removed: Tests that duplicate validator_test.go
// These tests were testing the underlying validation logic which is already
// thoroughly tested in validation/validator_test.go. The validator package
// is the source of truth for validation logic (host/domain/IP/CIDR matching).
//
// Kept: URL-layer tests (parsing, schemes, userinfo) that are unique to policy.go
// Kept: DNS resolution tests with go-mockdns (all tests below use mocked DNS)

func TestURLPolicyValidator_Validate_HostnameResolvesToIPInCIDR(t *testing.T) {
	// Test that hostname resolution + IP validation works correctly
	// Hostname resolves to IP that IS in allowed CIDR
	validator, srv, err := newValidatorWithMockDNS(
		config.WebhookSecurity{
			Mode:           config.WebhookSecurityModeCustom,
			AllowedSchemes: []string{"https"},
			AllowedHosts:   []string{"webhook.test"},
			AllowedCIDRs:   []string{"10.0.0.0/8"},
		},
		map[string]mockdns.Zone{
			"webhook.test.": {A: []string{"10.0.0.5"}}, // In 10.0.0.0/8
		},
	)
	require.NoError(t, err)
	defer srv.Close()

	err = validator.Validate(context.Background(), "https://webhook.test/hook")

	require.NoError(t, err)
}

func TestURLPolicyValidator_Validate_HostnameResolvesToIPNotInCIDR(t *testing.T) {
	// Test that hostname resolution + IP validation correctly rejects IPs not in CIDR
	// Hostname resolves to IP that is NOT in allowed CIDR
	validator, srv, err := newValidatorWithMockDNS(
		config.WebhookSecurity{
			Mode:           config.WebhookSecurityModeCustom,
			AllowedSchemes: []string{"https"},
			AllowedHosts:   []string{"webhook.test"},
			AllowedCIDRs:   []string{"192.168.1.0/24"},
		},
		map[string]mockdns.Zone{
			"webhook.test.": {A: []string{"10.0.0.5"}}, // NOT in 192.168.1.0/24
		},
	)
	require.NoError(t, err)
	defer srv.Close()

	err = validator.Validate(context.Background(), "https://webhook.test/hook")

	require.Error(t, err)
	require.Contains(t, err.Error(), "not in the allowed CIDR")
}

func TestURLPolicyValidator_Validate_HostnameResolvesAndReturnsMultipleIPs(t *testing.T) {
	// Test that ValidateAndGetIPs returns all validated IPs for a hostname
	validator, srv, err := newValidatorWithMockDNS(
		config.WebhookSecurity{
			Mode:           config.WebhookSecurityModeCustom,
			AllowedSchemes: []string{"https"},
			AllowedHosts:   []string{"webhook.test"},
			AllowedCIDRs:   []string{"10.0.0.0/8"},
		},
		map[string]mockdns.Zone{
			"webhook.test.": {
				A: []string{"10.0.0.5", "10.0.0.6"}, // Multiple IPs in allowed CIDR
			},
		},
	)
	require.NoError(t, err)
	defer srv.Close()

	result, err := validator.ValidateAndGetIPs(context.Background(), "https://webhook.test/hook")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.ValidatedIPs, 2)
	require.Equal(t, "10.0.0.5", result.ValidatedIPs[0].String())
	require.Equal(t, "10.0.0.6", result.ValidatedIPs[1].String())
	require.Equal(t, "webhook.test", result.Host)
}

func TestURLPolicyValidator_Validate_HostnameWithSkipResolvedIPValidation(t *testing.T) {
	// Test that SkipResolvedIPValidation allows hostname whose resolved IP is not in CIDR
	validator, srv, err := newValidatorWithMockDNS(
		config.WebhookSecurity{
			Mode:                     config.WebhookSecurityModeCustom,
			AllowedSchemes:           []string{"https"},
			AllowedHosts:             []string{"webhook.test"},
			AllowedCIDRs:             []string{"10.0.0.0/8"},
			SkipResolvedIPValidation: true,
		},
		map[string]mockdns.Zone{
			"webhook.test.": {A: []string{"8.8.8.8"}}, // NOT in 10.0.0.0/8, but skip flag is set
		},
	)
	require.NoError(t, err)
	defer srv.Close()

	err = validator.Validate(context.Background(), "https://webhook.test/hook")

	require.NoError(t, err)
}

func newValidatorWithMockDNS(security config.WebhookSecurity, zones map[string]mockdns.Zone) (*URLPolicyValidator, *mockdns.Server, error) {
	srv, err := mockdns.NewServer(zones, false)
	if err != nil {
		return nil, nil, err
	}

	validator := NewURLPolicyValidator(security)
	validator.resolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return net.Dial("udp", srv.LocalAddr().String())
		},
	}

	return validator, srv, nil
}

func TestURLPolicyValidator_DNSRebinding_RejectsPrivateIPInPublicOnlyMode(t *testing.T) {
	// Simulate an attacker controlling DNS to return a private IP
	validator, srv, err := newValidatorWithMockDNS(
		config.WebhookSecurity{
			Mode:           config.WebhookSecurityModePublicOnly,
			AllowedSchemes: []string{"https"},
		},
		map[string]mockdns.Zone{
			"evil.example.com.": {A: []string{"10.0.0.1"}}, // Private IP
		},
	)
	require.NoError(t, err)
	defer srv.Close()

	err = validator.Validate(context.Background(), "https://evil.example.com/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "public_only")
}

func TestURLPolicyValidator_DNSRebinding_RejectsMetadataIP(t *testing.T) {
	// Simulate DNS resolving to cloud metadata endpoint
	validator, srv, err := newValidatorWithMockDNS(
		config.WebhookSecurity{
			Mode:                  config.WebhookSecurityModeInsecure,
			AllowedSchemes:        []string{"https"},
			DenyMetadataEndpoints: true,
		},
		map[string]mockdns.Zone{
			"attacker.example.com.": {A: []string{"169.254.169.254"}}, // AWS metadata IP
		},
	)
	require.NoError(t, err)
	defer srv.Close()

	err = validator.Validate(context.Background(), "https://attacker.example.com/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "metadata")
}

func TestURLPolicyValidator_DNSRebinding_RejectsBlockedCIDR(t *testing.T) {
	// Simulate DNS resolving to a blocked CIDR range
	validator, srv, err := newValidatorWithMockDNS(
		config.WebhookSecurity{
			Mode:           config.WebhookSecurityModeCustom,
			AllowedSchemes: []string{"https"},
			AllowedHosts:   []string{"webhook.example.com"},
			AllowedCIDRs:   []string{"0.0.0.0/0"},
			BlockedCIDRs:   []string{"127.0.0.0/8"},
		},
		map[string]mockdns.Zone{
			"webhook.example.com.": {A: []string{"127.0.0.1"}}, // Blocked loopback
		},
	)
	require.NoError(t, err)
	defer srv.Close()

	err = validator.Validate(context.Background(), "https://webhook.example.com/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "blocked")
}

func TestURLPolicyValidator_DNSRebinding_RejectsIPNotInAllowedCIDR(t *testing.T) {
	// Simulate hostname resolving to IP outside allowed CIDR
	validator, srv, err := newValidatorWithMockDNS(
		config.WebhookSecurity{
			Mode:           config.WebhookSecurityModeCustom,
			AllowedSchemes: []string{"https"},
			AllowedHosts:   []string{"webhook.example.com"},
			AllowedCIDRs:   []string{"10.0.0.0/8"},
		},
		map[string]mockdns.Zone{
			"webhook.example.com.": {A: []string{"8.8.8.8"}}, // Not in 10.0.0.0/8
		},
	)
	require.NoError(t, err)
	defer srv.Close()

	err = validator.Validate(context.Background(), "https://webhook.example.com/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "not in the allowed CIDR")
}

func TestURLPolicyValidator_DNSRebinding_AllowsIPInAllowedCIDR(t *testing.T) {
	// Simulate hostname correctly resolving to allowed CIDR
	validator, srv, err := newValidatorWithMockDNS(
		config.WebhookSecurity{
			Mode:           config.WebhookSecurityModeCustom,
			AllowedSchemes: []string{"https"},
			AllowedHosts:   []string{"webhook.example.com"},
			AllowedCIDRs:   []string{"10.0.0.0/8"},
		},
		map[string]mockdns.Zone{
			"webhook.example.com.": {A: []string{"10.0.0.5"}}, // In 10.0.0.0/8
		},
	)
	require.NoError(t, err)
	defer srv.Close()

	result, err := validator.ValidateAndGetIPs(context.Background(), "https://webhook.example.com/webhook")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.ValidatedIPs, 1)
	require.Equal(t, "10.0.0.5", result.ValidatedIPs[0].String())
}

func TestURLPolicyValidator_DNSRebinding_RejectsIfAnyIPInvalid(t *testing.T) {
	// Simulate hostname resolving to multiple IPs where one is invalid
	// This tests that ALL resolved IPs must pass validation
	validator, srv, err := newValidatorWithMockDNS(
		config.WebhookSecurity{
			Mode:           config.WebhookSecurityModeCustom,
			AllowedSchemes: []string{"https"},
			AllowedHosts:   []string{"webhook.example.com"},
			AllowedCIDRs:   []string{"10.0.0.0/8"},
		},
		map[string]mockdns.Zone{
			"webhook.example.com.": {
				A: []string{
					"10.0.0.5", // Valid: in allowed CIDR
					"8.8.8.8",  // Invalid: not in allowed CIDR
				},
			},
		},
	)
	require.NoError(t, err)
	defer srv.Close()

	err = validator.Validate(context.Background(), "https://webhook.example.com/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "not in the allowed CIDR")
	require.ErrorContains(t, err, "8.8.8.8")
}

func TestURLPolicyValidator_DNSRebinding_AcceptsMultipleValidIPs(t *testing.T) {
	// Simulate hostname resolving to multiple valid IPs
	validator, srv, err := newValidatorWithMockDNS(
		config.WebhookSecurity{
			Mode:           config.WebhookSecurityModeCustom,
			AllowedSchemes: []string{"https"},
			AllowedHosts:   []string{"webhook.example.com"},
			AllowedCIDRs:   []string{"10.0.0.0/8"},
		},
		map[string]mockdns.Zone{
			"webhook.example.com.": {
				A: []string{"10.0.0.5", "10.0.0.6"},
			},
		},
	)
	require.NoError(t, err)
	defer srv.Close()

	result, err := validator.ValidateAndGetIPs(context.Background(), "https://webhook.example.com/webhook")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.ValidatedIPs, 2)
	require.Equal(t, "10.0.0.5", result.ValidatedIPs[0].String())
	require.Equal(t, "10.0.0.6", result.ValidatedIPs[1].String())
}

func TestURLPolicyValidator_DNSRebinding_PublicOnlyRejectsMixedIPs(t *testing.T) {
	// Simulate hostname resolving to both public and private IPs
	// This protects against attacks where DNS returns mixed public/internal IPs
	validator, srv, err := newValidatorWithMockDNS(
		config.WebhookSecurity{
			Mode:           config.WebhookSecurityModePublicOnly,
			AllowedSchemes: []string{"https"},
		},
		map[string]mockdns.Zone{
			"webhook.example.com.": {
				A: []string{
					"8.8.8.8",  // Public
					"10.0.0.1", // Private - should cause rejection
				},
			},
		},
	)
	require.NoError(t, err)
	defer srv.Close()

	err = validator.Validate(context.Background(), "https://webhook.example.com/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "public_only")
}

func TestURLPolicyValidator_DNSRebinding_SkipResolvedIPValidationStillRespectsBlocklist(t *testing.T) {
	// Even with SkipResolvedIPValidation=true, blocked CIDRs should still be enforced
	validator, srv, err := newValidatorWithMockDNS(
		config.WebhookSecurity{
			Mode:                     config.WebhookSecurityModeCustom,
			AllowedSchemes:           []string{"https"},
			AllowedHosts:             []string{"webhook.example.com"},
			SkipResolvedIPValidation: true,
			DenyMetadataEndpoints:    true,
		},
		map[string]mockdns.Zone{
			"webhook.example.com.": {A: []string{"169.254.169.254"}}, // Metadata IP
		},
	)
	require.NoError(t, err)
	defer srv.Close()

	err = validator.Validate(context.Background(), "https://webhook.example.com/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "metadata")
}

func TestURLPolicyValidator_DNSRebinding_InternalOnlyRejectsPublicIP(t *testing.T) {
	// Simulate hostname resolving to public IP in internal_only mode
	validator, srv, err := newValidatorWithMockDNS(
		config.WebhookSecurity{
			Mode:           config.WebhookSecurityModeInternalOnly,
			AllowedSchemes: []string{"https"},
		},
		map[string]mockdns.Zone{
			"webhook.internal.": {A: []string{"8.8.8.8"}}, // Public IP
		},
	)
	require.NoError(t, err)
	defer srv.Close()

	err = validator.Validate(context.Background(), "https://webhook.internal/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "internal_only")
}

func TestURLPolicyValidator_Validate_LiteralIPAddress(t *testing.T) {
	// Test that when host is already a literal IP, it's validated directly without DNS resolution
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeCustom,
		AllowedSchemes: []string{"https"},
		AllowedCIDRs:   []string{"10.0.0.0/8"},
	})

	result, err := validator.ValidateAndGetIPs(context.Background(), "https://10.0.0.5/webhook")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.ValidatedIPs, 1)
	require.Equal(t, "10.0.0.5", result.ValidatedIPs[0].String())
	require.Equal(t, "10.0.0.5", result.Host)
}

func TestURLPolicyValidator_Validate_LiteralIPAddress_Rejected(t *testing.T) {
	// Test that literal IP is rejected when not in allowed CIDR
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeCustom,
		AllowedSchemes: []string{"https"},
		AllowedCIDRs:   []string{"10.0.0.0/8"},
	})

	err := validator.Validate(context.Background(), "https://192.168.1.5/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "not in the allowed CIDR")
}

func TestURLPolicyValidator_Validate_HostnameResolvesToNoIPs(t *testing.T) {
	// Test that hostname that fails to resolve is rejected with appropriate error
	validator, srv, err := newValidatorWithMockDNS(
		config.WebhookSecurity{
			Mode:           config.WebhookSecurityModeCustom,
			AllowedSchemes: []string{"https"},
			AllowedHosts:   []string{"noips.test"},
		},
		map[string]mockdns.Zone{
			"noips.test.": {A: []string{}}, // No IPs returned - results in "no such host"
		},
	)
	require.NoError(t, err)
	defer srv.Close()

	err = validator.Validate(context.Background(), "https://noips.test/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to resolve webhook callback host")
}
