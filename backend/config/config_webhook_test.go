package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
)

func TestWebhooks_Decode(t *testing.T) {
	webhooks := Webhooks{}
	value := "{\"callback\":\"http://app.com/usercb\",\"events\":[\"user\"]};{\"callback\":\"http://app.com/callback\",\"events\":[\"email.send\"]}"

	err := webhooks.Decode(value)

	assert.NoError(t, err)
	assert.Len(t, webhooks, 2)
	for _, webhook := range webhooks {
		assert.IsType(t, Webhook{}, webhook)
	}
}

func TestWebhookSecurity_Validate_AcceptsValidSecurityConfigWithAllowlist(t *testing.T) {
	security := WebhookSecurity{
		Mode:                  WebhookSecurityModeCustom,
		AllowedSchemes:        []string{"http", "https"},
		FollowRedirects:       true,
		MaxRedirects:          3,
		AllowedHosts:          []string{"localhost"},
		AllowedDomains:        []string{"example.com"},
		AllowedCIDRs:          []string{"127.0.0.0/8", "10.0.0.0/24"},
		DenyMetadataEndpoints: true,
	}

	err := security.Validate()

	assert.NoError(t, err)
}

func TestWebhookSecurity_Validate_AcceptsValidSecurityConfigWithBlocklist(t *testing.T) {
	security := WebhookSecurity{
		Mode:                  WebhookSecurityModeCustom,
		AllowedSchemes:        []string{"http", "https"},
		AllowedHosts:          []string{"example.com"},
		FollowRedirects:       true,
		MaxRedirects:          3,
		BlockedDomains:        []string{"internal.example"},
		BlockedCIDRs:          []string{"169.254.169.254/32"},
		DenyMetadataEndpoints: true,
	}

	err := security.Validate()

	assert.NoError(t, err)
}

func TestWebhookSecurity_Validate_RejectsInvalidMode(t *testing.T) {
	security := WebhookSecurity{
		Mode: "nope",
	}

	err := security.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhooks.security.mode")
}

func TestWebhookSecurity_Validate_RejectsInvalidScheme(t *testing.T) {
	security := WebhookSecurity{
		Mode:           WebhookSecurityModePublicOnly,
		AllowedSchemes: []string{"ftp"},
	}

	err := security.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "allowed_schemes")
}

func TestWebhookSecurity_Validate_RejectsInvalidRedirectConfiguration(t *testing.T) {
	security := WebhookSecurity{
		Mode:            WebhookSecurityModePublicOnly,
		AllowedSchemes:  []string{"https"},
		FollowRedirects: false,
		MaxRedirects:    1,
	}

	err := security.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max_redirects")
}

func TestWebhookSecurity_Validate_RejectsNegativeRedirectCount(t *testing.T) {
	security := WebhookSecurity{
		Mode:            WebhookSecurityModePublicOnly,
		AllowedSchemes:  []string{"https"},
		FollowRedirects: true,
		MaxRedirects:    -1,
	}

	err := security.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "greater than or equal to 0")
}

func TestWebhookSecurity_Validate_RejectsInvalidAllowedCIDR(t *testing.T) {
	security := WebhookSecurity{
		Mode:         WebhookSecurityModeCustom,
		AllowedCIDRs: []string{"definitely-not-a-cidr"},
	}

	err := security.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "allowed_cidrs")
}

func TestWebhookSecurity_Validate_RejectsInvalidBlockedCIDR(t *testing.T) {
	security := WebhookSecurity{
		Mode:         WebhookSecurityModeCustom,
		AllowedCIDRs: []string{"10.0.0.0/24"},
		BlockedCIDRs: []string{"still-not-a-cidr"},
	}

	err := security.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blocked_cidrs")
}

func TestWebhookSecurity_Validate_RejectsEmptyAllowedHost(t *testing.T) {
	security := WebhookSecurity{
		Mode:         WebhookSecurityModeCustom,
		AllowedHosts: []string{""},
	}

	err := security.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "allowed_hosts")
}

func TestWebhookSecurity_Validate_RejectsEmptyBlockedHost(t *testing.T) {
	security := WebhookSecurity{
		Mode:         WebhookSecurityModeCustom,
		AllowedHosts: []string{"example.com"},
		BlockedHosts: []string{""},
	}

	err := security.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blocked_hosts")
}

func TestWebhookSecurity_Validate_RejectsEmptyAllowedDomain(t *testing.T) {
	security := WebhookSecurity{
		Mode:           WebhookSecurityModeCustom,
		AllowedDomains: []string{""},
	}

	err := security.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "allowed_domains")
}

func TestWebhookSecurity_Validate_RejectsAllowedDomainWithPort(t *testing.T) {
	security := WebhookSecurity{
		Mode:           WebhookSecurityModeCustom,
		AllowedDomains: []string{"example.com:443"},
	}

	err := security.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must not contain a port")
}

func TestWebhookSecurity_Validate_RejectsBlockedDomainWithPort(t *testing.T) {
	security := WebhookSecurity{
		Mode:           WebhookSecurityModeCustom,
		AllowedHosts:   []string{"allowed.com"},
		BlockedDomains: []string{"example.com:443"},
	}

	err := security.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must not contain a port")
}

func TestWebhookSettings_Validate_DisabledSkipsValidation(t *testing.T) {
	settings := WebhookSettings{
		Enabled: false,
		Security: WebhookSecurity{
			Mode: "definitely-invalid",
		},
		Hooks: Webhooks{
			{
				Callback: "://broken",
				Events:   events.Events{events.Event("not-an-event")},
			},
		},
	}

	err := settings.Validate()

	assert.NoError(t, err)
}

// Removed: URL validation tests that duplicate validator_test.go
// These tests validate callback URLs through WebhookSettings.Validate() which calls
// validation.ValidateWebhook(), duplicating tests already in validator_test.go.
// The validator_test.go file is the source of truth for validation logic.
//
// Integration testing of webhook validation is done in policy_test.go.

func TestWebhookSecurity_Validate_CustomModeRequiresAtLeastOneAllowlist(t *testing.T) {
	security := WebhookSecurity{
		Mode:           WebhookSecurityModeCustom,
		AllowedSchemes: []string{"https"},
		// No allowlists configured
	}

	err := security.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires at least one allow list")
}

func TestWebhookSecurity_Validate_CustomModeAcceptsAllowedHosts(t *testing.T) {
	security := WebhookSecurity{
		Mode:           WebhookSecurityModeCustom,
		AllowedSchemes: []string{"https"},
		AllowedHosts:   []string{"example.com"},
	}

	err := security.Validate()

	assert.NoError(t, err)
}

func TestWebhookSecurity_Validate_CustomModeAcceptsAllowedDomains(t *testing.T) {
	security := WebhookSecurity{
		Mode:           WebhookSecurityModeCustom,
		AllowedSchemes: []string{"https"},
		AllowedDomains: []string{"example.com"},
	}

	err := security.Validate()

	assert.NoError(t, err)
}

func TestWebhookSecurity_Validate_CustomModeAcceptsAllowedCIDRs(t *testing.T) {
	security := WebhookSecurity{
		Mode:           WebhookSecurityModeCustom,
		AllowedSchemes: []string{"https"},
		AllowedCIDRs:   []string{"10.0.0.0/24"},
	}

	err := security.Validate()

	assert.NoError(t, err)
}

func TestWebhookSettings_Validate_RejectsInvalidEvent(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModeInsecure,
			AllowedSchemes: []string{"http", "https"},
		},
		Hooks: Webhooks{
			{
				Callback: "http://example.com/webhook",
				Events:   events.Events{events.Event("not-a-valid-event")},
			},
		},
	}

	err := settings.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a valid webhook event")
}

// Integration test to ensure callback URL validation is wired up correctly
func TestWebhookSettings_Validate_IntegrationCallbackValidation(t *testing.T) {
	// Test that webhook validation is actually called during settings validation
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModePublicOnly,
			AllowedSchemes: []string{"https"},
		},
		Hooks: Webhooks{
			{
				Callback: "https://example.com/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	// This should pass - the integration between WebhookSettings and validation package works
	assert.NoError(t, err)
}

func TestWebhookSecurity_Validate_AllowsPublicOnlyWithIgnoredConfigs(t *testing.T) {
	security := WebhookSecurity{
		Mode:           WebhookSecurityModePublicOnly,
		AllowedSchemes: []string{"https"},
		AllowedHosts:   []string{"example.com"},
		BlockedDomains: []string{"blocked.com"},
		AllowedCIDRs:   []string{"10.0.0.0/8"},
	}

	err := security.Validate()

	// Should not return an error - validation should pass
	// (warnings will be logged but validation succeeds)
	assert.NoError(t, err)
}

func TestWebhookSecurity_Validate_AllowsInternalOnlyWithIgnoredConfigs(t *testing.T) {
	security := WebhookSecurity{
		Mode:           WebhookSecurityModeInternalOnly,
		AllowedSchemes: []string{"http", "https"},
		AllowedCIDRs:   []string{"10.0.0.0/8"},
		BlockedHosts:   []string{"blocked.example.com"},
		AllowedDomains: []string{"example.com"},
	}

	err := security.Validate()

	// Should not return an error - validation should pass
	// (warnings will be logged but validation succeeds)
	assert.NoError(t, err)
}

func TestWebhookSecurity_Validate_RejectsCustomWithBothAllowedAndBlockedCIDRs(t *testing.T) {
	security := WebhookSecurity{
		Mode:           WebhookSecurityModeCustom,
		AllowedSchemes: []string{"https"},
		AllowedCIDRs:   []string{"10.0.0.0/24"},
		BlockedCIDRs:   []string{"192.168.0.0/24"},
	}

	err := security.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
	assert.Contains(t, err.Error(), "allowed_cidrs")
	assert.Contains(t, err.Error(), "blocked_cidrs")
}

func TestWebhookSecurity_Validate_RejectsCustomWithBothAllowedAndBlockedHosts(t *testing.T) {
	security := WebhookSecurity{
		Mode:           WebhookSecurityModeCustom,
		AllowedSchemes: []string{"https"},
		AllowedHosts:   []string{"example.com"},
		BlockedHosts:   []string{"blocked.com"},
	}

	err := security.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
	assert.Contains(t, err.Error(), "allowed_hosts")
	assert.Contains(t, err.Error(), "blocked_hosts")
}

func TestWebhookSecurity_Validate_RejectsCustomWithBothAllowedAndBlockedDomains(t *testing.T) {
	security := WebhookSecurity{
		Mode:           WebhookSecurityModeCustom,
		AllowedSchemes: []string{"https"},
		AllowedDomains: []string{"example.com"},
		BlockedDomains: []string{"blocked.com"},
	}

	err := security.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
	assert.Contains(t, err.Error(), "allowed_domains")
	assert.Contains(t, err.Error(), "blocked_domains")
}

func TestWebhookSecurity_Validate_AllowsPublicOnlyWithOnlyAllowedSchemes(t *testing.T) {
	security := WebhookSecurity{
		Mode:           WebhookSecurityModePublicOnly,
		AllowedSchemes: []string{"https"},
	}

	err := security.Validate()

	assert.NoError(t, err)
}

func TestWebhookSecurity_Validate_AllowsInternalOnlyWithOnlyAllowedSchemes(t *testing.T) {
	security := WebhookSecurity{
		Mode:           WebhookSecurityModeInternalOnly,
		AllowedSchemes: []string{"http", "https"},
	}

	err := security.Validate()

	assert.NoError(t, err)
}
