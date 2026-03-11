package webhooks

import (
	"context"
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/v2/config"
)

func TestNormalizePolicyHost(t *testing.T) {
	require.Equal(t, "example.com", normalizePolicyHost(" Example.com. "))
}

func TestMatchesHost(t *testing.T) {
	require.True(t, matchesHost("Example.com", []string{"example.com"}))
	require.False(t, matchesHost("api.example.com", []string{"example.com"}))
}

func TestMatchesDomain(t *testing.T) {
	require.True(t, matchesDomain("example.com", []string{"example.com"}))
	require.True(t, matchesDomain("api.example.com", []string{"example.com"}))
	require.True(t, matchesDomain("a.b.example.com", []string{"example.com"}))
	require.False(t, matchesDomain("badexample.com", []string{"example.com"}))
}

func TestIPMatchesCIDRs(t *testing.T) {
	require.True(t, ipMatchesCIDRs(net.ParseIP("10.0.0.5"), []string{"10.0.0.0/24"}))
	require.False(t, ipMatchesCIDRs(net.ParseIP("10.0.1.5"), []string{"10.0.0.0/24"}))
}

func TestIsMetadataIP(t *testing.T) {
	require.True(t, isMetadataIP(net.ParseIP("169.254.169.254")))
	require.False(t, isMetadataIP(net.ParseIP("169.254.169.253")))
}

func TestIsPublicRoutableIP(t *testing.T) {
	require.True(t, isPublicRoutableIP(net.ParseIP("8.8.8.8")))
	require.False(t, isPublicRoutableIP(net.ParseIP("127.0.0.1")))
	require.False(t, isPublicRoutableIP(net.ParseIP("10.0.0.1")))
	require.False(t, isPublicRoutableIP(net.ParseIP("169.254.169.254")))
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
