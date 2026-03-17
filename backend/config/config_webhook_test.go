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
		FollowRedirects:       true,
		MaxRedirects:          3,
		BlockedHosts:          []string{"metadata.google.internal"},
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

func TestWebhookSettings_Validate_PublicOnlyRejectsHTTPCallbackWhenSchemeNotAllowed(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModePublicOnly,
			AllowedSchemes: []string{"https"},
		},
		Hooks: Webhooks{
			{
				Callback: "http://example.com/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scheme")
}

func TestWebhookSettings_Validate_InsecureStillRejectsDisallowedScheme(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModeInsecure,
			AllowedSchemes: []string{"https"},
		},
		Hooks: Webhooks{
			{
				Callback: "http://example.com/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scheme")
}

func TestWebhookSettings_Validate_InsecureAllowsHTTPWhenConfigured(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModeInsecure,
			AllowedSchemes: []string{"http", "https"},
		},
		Hooks: Webhooks{
			{
				Callback: "http://example.com/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.NoError(t, err)
}

func TestWebhookSettings_Validate_RejectsBlockedHost(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModeCustom,
			AllowedSchemes: []string{"http", "https"},
			BlockedHosts:   []string{"api.example.com"},
		},
		Hooks: Webhooks{
			{
				Callback: "https://api.example.com/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blocked")
}

func TestWebhookSettings_Validate_RejectsBlockedDomain(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModeCustom,
			AllowedSchemes: []string{"http", "https"},
			BlockedDomains: []string{"example.com"},
		},
		Hooks: Webhooks{
			{
				Callback: "https://api.example.com/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blocked domain")
}

func TestWebhookSettings_Validate_CustomAllowsHostnameWhenAllowedHostMatches(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModeCustom,
			AllowedSchemes: []string{"https"},
			AllowedHosts:   []string{"api.example.com"},
		},
		Hooks: Webhooks{
			{
				Callback: "https://api.example.com/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.NoError(t, err)
}

func TestWebhookSettings_Validate_CustomAllowsHostnameWhenAllowedDomainMatches(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModeCustom,
			AllowedSchemes: []string{"https"},
			AllowedDomains: []string{"example.com"},
		},
		Hooks: Webhooks{
			{
				Callback: "https://api.example.com/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.NoError(t, err)
}

func TestWebhookSettings_Validate_CustomRejectsHostnameWhenAllowedListsExistButNoMatch(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModeCustom,
			AllowedSchemes: []string{"https"},
			AllowedDomains: []string{"example.com"},
		},
		Hooks: Webhooks{
			{
				Callback: "https://api.other.com/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "allowed host/domain list")
}

func TestWebhookSettings_Validate_CustomRejectsLiteralPrivateIPWithoutAllowedHostOrCIDR(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModeCustom,
			AllowedSchemes: []string{"http", "https"},
		},
		Hooks: Webhooks{
			{
				Callback: "http://10.0.0.2/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "explicitly allowlisted")
}

func TestWebhookSettings_Validate_CustomRejectsLiteralPrivateIPWhenHostAllowedButCIDRNotAllowed(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModeCustom,
			AllowedSchemes: []string{"http", "https"},
			AllowedHosts:   []string{"10.0.0.2"},
		},
		Hooks: Webhooks{
			{
				Callback: "http://10.0.0.2/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "explicitly allowlisted")
}

func TestWebhookSettings_Validate_CustomAllowsLiteralPrivateIPWhenHostAndCIDRAreAllowed(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModeCustom,
			AllowedSchemes: []string{"http", "https"},
			AllowedHosts:   []string{"10.0.0.2"},
			AllowedCIDRs:   []string{"10.0.0.0/24"},
		},
		Hooks: Webhooks{
			{
				Callback: "http://10.0.0.2/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.NoError(t, err)
}

func TestWebhookSettings_Validate_PublicOnlyRejectsLiteralPrivateIP(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModePublicOnly,
			AllowedSchemes: []string{"http", "https"},
		},
		Hooks: Webhooks{
			{
				Callback: "http://10.0.0.2/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "public_only")
}

func TestWebhookSettings_Validate_InternalOnlyAllowsLiteralPrivateIP(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModeInternalOnly,
			AllowedSchemes: []string{"http", "https"},
		},
		Hooks: Webhooks{
			{
				Callback: "http://10.0.0.2/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.NoError(t, err)
}

func TestWebhookSettings_Validate_InternalOnlyRejectsPublicIP(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModeInternalOnly,
			AllowedSchemes: []string{"http", "https"},
		},
		Hooks: Webhooks{
			{
				Callback: "http://8.8.8.8/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "internal_only")
}

func TestWebhookSettings_Validate_CustomAllowsLiteralPrivateIPWhenCIDRAllowlisted(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModeCustom,
			AllowedSchemes: []string{"http", "https"},
			AllowedCIDRs:   []string{"10.0.0.0/24"},
		},
		Hooks: Webhooks{
			{
				Callback: "http://10.0.0.2/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.NoError(t, err)
}

func TestWebhookSettings_Validate_RejectsBlockedLiteralIP(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:           WebhookSecurityModeCustom,
			AllowedSchemes: []string{"http", "https"},
			BlockedCIDRs:   []string{"127.0.0.0/8"},
		},
		Hooks: Webhooks{
			{
				Callback: "http://127.0.0.1/webhook",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blocked")
}

func TestWebhookSettings_Validate_RejectsMetadataEndpointIP(t *testing.T) {
	settings := WebhookSettings{
		Enabled: true,
		Security: WebhookSecurity{
			Mode:                  WebhookSecurityModeCustom,
			AllowedSchemes:        []string{"http", "https"},
			DenyMetadataEndpoints: true,
		},
		Hooks: Webhooks{
			{
				Callback: "http://169.254.169.254/latest/meta-data",
				Events:   events.Events{events.Event("user.create")},
			},
		},
	}

	err := settings.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metadata")
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
