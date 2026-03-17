package validation

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidateHost_RejectsEmptyHost(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode: SecurityModePublicOnly,
	})

	err := validator.ValidateHost("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must not be empty")
}

func TestValidator_ValidateHost_RejectsBlockedHost(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode:         SecurityModePublicOnly,
		BlockedHosts: []string{"blocked.example.com"},
	})

	err := validator.ValidateHost("blocked.example.com")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is blocked")
}

func TestValidator_ValidateHost_RejectsBlockedDomain(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode:           SecurityModePublicOnly,
		BlockedDomains: []string{"example.com"},
	})

	err := validator.ValidateHost("api.example.com")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blocked domain")
}

func TestValidator_ValidateHost_RejectsMetadataHost(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode:                  SecurityModePublicOnly,
		DenyMetadataEndpoints: true,
	})

	err := validator.ValidateHost("metadata.google.internal")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metadata")
}

func TestValidator_ValidateHost_CustomModeAllowsHostInAllowlist(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode:         SecurityModeCustom,
		AllowedHosts: []string{"allowed.example.com"},
	})

	err := validator.ValidateHost("allowed.example.com")

	assert.NoError(t, err)
}

func TestValidator_ValidateHost_CustomModeAllowsHostInAllowedDomain(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode:           SecurityModeCustom,
		AllowedDomains: []string{"example.com"},
	})

	err := validator.ValidateHost("api.example.com")

	assert.NoError(t, err)
}

func TestValidator_ValidateHost_CustomModeRejectsHostNotInAllowlist(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode:           SecurityModeCustom,
		AllowedDomains: []string{"example.com"},
	})

	err := validator.ValidateHost("other.com")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in the allowed")
}

func TestValidator_ValidateIP_RejectsBlockedCIDR(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode:         SecurityModePublicOnly,
		BlockedCIDRs: []string{"10.0.0.0/24"},
	})

	err := validator.ValidateIP(net.ParseIP("10.0.0.5"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blocked CIDR")
}

func TestValidator_ValidateIP_RejectsMetadataIP(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode:                  SecurityModePublicOnly,
		DenyMetadataEndpoints: true,
	})

	err := validator.ValidateIP(net.ParseIP("169.254.169.254"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metadata")
}

func TestValidator_ValidateIP_PublicOnlyRejectsPrivateIP(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode: SecurityModePublicOnly,
	})

	err := validator.ValidateIP(net.ParseIP("10.0.0.1"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "public_only")
}

func TestValidator_ValidateIP_PublicOnlyAllowsPublicIP(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode: SecurityModePublicOnly,
	})

	err := validator.ValidateIP(net.ParseIP("8.8.8.8"))

	assert.NoError(t, err)
}

func TestValidator_ValidateIP_InternalOnlyAllowsPrivateIP(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode: SecurityModeInternalOnly,
	})

	err := validator.ValidateIP(net.ParseIP("10.0.0.1"))

	assert.NoError(t, err)
}

func TestValidator_ValidateIP_InternalOnlyRejectsPublicIP(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode: SecurityModeInternalOnly,
	})

	err := validator.ValidateIP(net.ParseIP("8.8.8.8"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "internal_only")
}

func TestValidator_ValidateIP_CustomAllowsPrivateIPInAllowedCIDR(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode:         SecurityModeCustom,
		AllowedCIDRs: []string{"10.0.0.0/24"},
	})

	err := validator.ValidateIP(net.ParseIP("10.0.0.5"))

	assert.NoError(t, err)
}

func TestValidator_ValidateIP_CustomRejectsPrivateIPNotInAllowedCIDR(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode:         SecurityModeCustom,
		AllowedCIDRs: []string{"10.0.0.0/24"},
	})

	err := validator.ValidateIP(net.ParseIP("10.0.1.5"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "explicitly allowlisted")
}

func TestValidator_ValidateIP_InsecureAllowsAnyIP(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode: SecurityModeInsecure,
	})

	err := validator.ValidateIP(net.ParseIP("127.0.0.1"))

	assert.NoError(t, err)
}

func TestValidator_ValidateIP_InsecureStillRespectsBlockedCIDRs(t *testing.T) {
	validator := NewValidator(WebhookSecurityPolicy{
		Mode:         SecurityModeInsecure,
		BlockedCIDRs: []string{"127.0.0.0/8"},
	})

	err := validator.ValidateIP(net.ParseIP("127.0.0.1"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blocked")
}
