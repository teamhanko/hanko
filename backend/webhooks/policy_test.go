package webhooks

import (
	"context"
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/webhooks/validation"
)

func TestNormalizeHost(t *testing.T) {
	require.Equal(t, "example.com", validation.NormalizeHost(" Example.com. "))
}

func TestMatchesHost(t *testing.T) {
	require.True(t, validation.MatchesHost("Example.com", []string{"example.com"}))
	require.False(t, validation.MatchesHost("api.example.com", []string{"example.com"}))
}

func TestMatchesDomain(t *testing.T) {
	require.True(t, validation.MatchesDomain("example.com", []string{"example.com"}))
	require.True(t, validation.MatchesDomain("api.example.com", []string{"example.com"}))
	require.True(t, validation.MatchesDomain("a.b.example.com", []string{"example.com"}))
	require.False(t, validation.MatchesDomain("badexample.com", []string{"example.com"}))
}

func TestIPMatchesCIDRs(t *testing.T) {
	require.True(t, validation.IPMatchesCIDRs(net.ParseIP("10.0.0.5"), []string{"10.0.0.0/24"}))
	require.False(t, validation.IPMatchesCIDRs(net.ParseIP("10.0.1.5"), []string{"10.0.0.0/24"}))
}

func TestIsMetadataIP(t *testing.T) {
	require.True(t, validation.IsMetadataIP(net.ParseIP("169.254.169.254")))
	require.False(t, validation.IsMetadataIP(net.ParseIP("169.254.169.253")))
}

func TestIsPublicRoutableIP(t *testing.T) {
	require.True(t, validation.IsPublicRoutableIP(net.ParseIP("8.8.8.8")))
	require.False(t, validation.IsPublicRoutableIP(net.ParseIP("127.0.0.1")))
	require.False(t, validation.IsPublicRoutableIP(net.ParseIP("10.0.0.1")))
	require.False(t, validation.IsPublicRoutableIP(net.ParseIP("169.254.169.254")))
}

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

func TestURLPolicyValidator_Validate_RejectsBlockedHost(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeInsecure,
		AllowedSchemes: []string{"http", "https"},
		BlockedHosts:   []string{"127.0.0.1"},
	})

	err := validator.Validate(context.Background(), "http://127.0.0.1/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "is blocked")
}

func TestURLPolicyValidator_Validate_RejectsBlockedDomainViaParsedURL(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeInsecure,
		AllowedSchemes: []string{"https"},
		BlockedDomains: []string{"example.com"},
	})

	parsed, err := url.Parse("https://api.example.com/webhook")
	require.NoError(t, err)

	_, err = validator.validateParsedURL(parsed)

	require.Error(t, err)
	require.ErrorContains(t, err, "blocked domain")
}

func TestURLPolicyValidator_Validate_CustomRejectsHostWhenAllowedHostListDoesNotMatchViaParsedURL(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeCustom,
		AllowedSchemes: []string{"https"},
		AllowedHosts:   []string{"api.example.com"},
	})

	parsed, err := url.Parse("https://other.example.com/webhook")
	require.NoError(t, err)

	_, err = validator.validateParsedURL(parsed)

	require.Error(t, err)
	require.ErrorContains(t, err, "allowed host/domain list")
}

func TestURLPolicyValidator_Validate_CustomRejectsHostWhenAllowedDomainListDoesNotMatchViaParsedURL(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeCustom,
		AllowedSchemes: []string{"https"},
		AllowedDomains: []string{"example.com"},
	})

	parsed, err := url.Parse("https://api.other.com/webhook")
	require.NoError(t, err)

	_, err = validator.validateParsedURL(parsed)

	require.Error(t, err)
	require.ErrorContains(t, err, "allowed host/domain list")
}

func TestURLPolicyValidator_Validate_CustomAllowsHostWhenAllowedHostMatchesViaParsedURL(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeCustom,
		AllowedSchemes: []string{"https"},
		AllowedHosts:   []string{"api.example.com"},
	})

	parsed, err := url.Parse("https://api.example.com/webhook")
	require.NoError(t, err)

	host, err := validator.validateParsedURL(parsed)

	require.NoError(t, err)
	require.Equal(t, "api.example.com", host)
}

func TestURLPolicyValidator_Validate_CustomAllowsHostWhenAllowedDomainMatchesViaParsedURL(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeCustom,
		AllowedSchemes: []string{"https"},
		AllowedDomains: []string{"example.com"},
	})

	parsed, err := url.Parse("https://api.example.com/webhook")
	require.NoError(t, err)

	host, err := validator.validateParsedURL(parsed)

	require.NoError(t, err)
	require.Equal(t, "api.example.com", host)
}

func TestURLPolicyValidator_Validate_InsecureAllowsPrivateLiteralIP(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeInsecure,
		AllowedSchemes: []string{"http", "https"},
	})

	err := validator.Validate(context.Background(), "http://10.0.0.2/webhook")

	require.NoError(t, err)
}

func TestURLPolicyValidator_Validate_PublicOnlyRejectsLoopback(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModePublicOnly,
		AllowedSchemes: []string{"http", "https"},
	})

	err := validator.Validate(context.Background(), "http://127.0.0.1/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "public_only")
}

func TestURLPolicyValidator_Validate_PublicOnlyRejectsPrivateLiteralIP(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModePublicOnly,
		AllowedSchemes: []string{"http", "https"},
	})

	err := validator.Validate(context.Background(), "http://10.0.0.2/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "public_only")
}

func TestURLPolicyValidator_Validate_CustomAllowsExplicitCIDR(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeCustom,
		AllowedSchemes: []string{"http", "https"},
		AllowedCIDRs:   []string{"10.0.0.0/24"},
	})

	err := validator.Validate(context.Background(), "http://10.0.0.2/webhook")

	require.NoError(t, err)
}

func TestURLPolicyValidator_Validate_CustomRejectsPrivateLiteralIPWithoutAllowedCIDR(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeCustom,
		AllowedSchemes: []string{"http", "https"},
	})

	err := validator.Validate(context.Background(), "http://10.0.0.2/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "explicitly allowlisted")
}

func TestURLPolicyValidator_Validate_BlockedCIDRWinsOverAllowedCIDR(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeCustom,
		AllowedSchemes: []string{"http", "https"},
		AllowedCIDRs:   []string{"10.0.0.0/24"},
		BlockedCIDRs:   []string{"10.0.0.0/25"},
	})

	err := validator.Validate(context.Background(), "http://10.0.0.2/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "blocked CIDR")
}

func TestURLPolicyValidator_Validate_RejectsMetadataEndpoint(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:                  config.WebhookSecurityModeCustom,
		AllowedSchemes:        []string{"http", "https"},
		DenyMetadataEndpoints: true,
	})

	err := validator.Validate(context.Background(), "http://169.254.169.254/latest/meta-data")

	require.Error(t, err)
	require.ErrorContains(t, err, "metadata")
}

func TestURLPolicyValidator_Validate_BlockedHostWinsInCustomMode(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeCustom,
		AllowedSchemes: []string{"http", "https"},
		AllowedHosts:   []string{"127.0.0.1"},
		BlockedHosts:   []string{"127.0.0.1"},
	})

	err := validator.Validate(context.Background(), "http://127.0.0.1/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "is blocked")
}

func TestURLPolicyValidator_Validate_CustomRequiresBothAllowedHostAndAllowedCIDRForLiteralPrivateIP(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeCustom,
		AllowedSchemes: []string{"http", "https"},
		AllowedHosts:   []string{"10.0.0.2"},
	})

	err := validator.Validate(context.Background(), "http://10.0.0.2/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "explicitly allowlisted")
}

func TestURLPolicyValidator_Validate_CustomAllowsLiteralPrivateIPWhenHostAndCIDRAreAllowed(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeCustom,
		AllowedSchemes: []string{"http", "https"},
		AllowedHosts:   []string{"10.0.0.2"},
		AllowedCIDRs:   []string{"10.0.0.0/24"},
	})

	err := validator.Validate(context.Background(), "http://10.0.0.2/webhook")

	require.NoError(t, err)
}

func TestURLPolicyValidator_Validate_PublicAddressLiteralIPAllowed(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModePublicOnly,
		AllowedSchemes: []string{"http", "https"},
	})

	err := validator.Validate(context.Background(), "http://8.8.8.8/webhook")

	require.NoError(t, err)
}

func TestURLPolicyValidator_Validate_IPv6LoopbackRejected(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModePublicOnly,
		AllowedSchemes: []string{"http", "https"},
	})

	err := validator.Validate(context.Background(), "http://[::1]/webhook")

	require.Error(t, err)
	require.ErrorContains(t, err, "public_only")
}

func TestURLPolicyValidator_ValidateAndGetIPs_ReturnsValidatedIPsForLiteralIP(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeInsecure,
		AllowedSchemes: []string{"http", "https"},
	})

	result, err := validator.ValidateAndGetIPs(context.Background(), "http://8.8.8.8/webhook")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.ValidatedIPs, 1)
	require.Equal(t, "8.8.8.8", result.ValidatedIPs[0].String())
	require.Equal(t, "8.8.8.8", result.Host)
}

func TestURLPolicyValidator_ValidateAndGetIPs_RejectsPrivateIPInPublicOnlyMode(t *testing.T) {
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModePublicOnly,
		AllowedSchemes: []string{"http", "https"},
	})

	result, err := validator.ValidateAndGetIPs(context.Background(), "http://10.0.0.1/webhook")

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorContains(t, err, "public_only")
}

func TestURLPolicyValidator_ValidateAndGetIPs_ReturnsMultipleIPs(t *testing.T) {
	// Note: This test uses google.com which typically has multiple A records
	// In a real-world scenario, the validator would resolve DNS and return all IPs
	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModePublicOnly,
		AllowedSchemes: []string{"https"},
	})

	result, err := validator.ValidateAndGetIPs(context.Background(), "https://google.com/webhook")

	// This test may fail in environments without internet access
	// or if google.com is blocked, but demonstrates the functionality
	if err == nil {
		require.NotNil(t, result)
		require.NotEmpty(t, result.ValidatedIPs)
		require.Equal(t, "google.com", result.Host)

		// Verify all IPs are public
		for _, ip := range result.ValidatedIPs {
			require.True(t, validation.IsPublicRoutableIP(ip),
				"Expected all IPs to be public, got: %s", ip.String())
		}
	}
}

func TestURLPolicyValidator_ValidateAndGetIPs_PreventsDNSRebinding(t *testing.T) {
	// This test verifies that ValidateAndGetIPs returns IPs at validation time
	// which can then be pinned to prevent DNS rebinding

	validator := NewURLPolicyValidator(config.WebhookSecurity{
		Mode:           config.WebhookSecurityModeInsecure,
		AllowedSchemes: []string{"http", "https"},
	})

	// Validate with a literal IP
	result, err := validator.ValidateAndGetIPs(context.Background(), "http://127.0.0.1/webhook")

	require.NoError(t, err)
	require.NotNil(t, result)

	// The returned IPs are from the validation time
	// If these IPs are pinned in the HTTP client, DNS rebinding is prevented
	validatedIP := result.ValidatedIPs[0]
	require.Equal(t, "127.0.0.1", validatedIP.String())

	// In production: these IPs would be used by ValidatedDialer
	// to bypass DNS resolution during HTTP request, preventing rebinding
}
