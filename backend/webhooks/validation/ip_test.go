package validation

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPMatchesCIDRs_MatchesValidCIDR(t *testing.T) {
	ip := net.ParseIP("10.0.0.5")
	cidrs := []string{"10.0.0.0/24"}

	result := IPMatchesCIDRs(ip, cidrs)

	assert.True(t, result)
}

func TestIPMatchesCIDRs_DoesNotMatchDifferentCIDR(t *testing.T) {
	ip := net.ParseIP("10.0.1.5")
	cidrs := []string{"10.0.0.0/24"}

	result := IPMatchesCIDRs(ip, cidrs)

	assert.False(t, result)
}

func TestIPMatchesCIDRs_IgnoresInvalidCIDR(t *testing.T) {
	ip := net.ParseIP("10.0.0.5")
	cidrs := []string{"not-a-cidr", "10.0.0.0/24"}

	result := IPMatchesCIDRs(ip, cidrs)

	assert.True(t, result)
}

func TestIPMatchesCIDRs_HandlesIPv6(t *testing.T) {
	ip := net.ParseIP("2001:db8::1")
	cidrs := []string{"2001:db8::/32"}

	result := IPMatchesCIDRs(ip, cidrs)

	assert.True(t, result)
}

func TestIsPublicRoutableIP_ReturnsTrueForPublicIP(t *testing.T) {
	ip := net.ParseIP("8.8.8.8")

	result := IsPublicRoutableIP(ip)

	assert.True(t, result)
}

func TestIsPublicRoutableIP_ReturnsFalseForLoopback(t *testing.T) {
	ip := net.ParseIP("127.0.0.1")

	result := IsPublicRoutableIP(ip)

	assert.False(t, result)
}

func TestIsPublicRoutableIP_ReturnsFalseForPrivate(t *testing.T) {
	ip := net.ParseIP("10.0.0.1")

	result := IsPublicRoutableIP(ip)

	assert.False(t, result)
}

func TestIsPublicRoutableIP_ReturnsFalseForMulticast(t *testing.T) {
	ip := net.ParseIP("224.0.0.1")

	result := IsPublicRoutableIP(ip)

	assert.False(t, result)
}

func TestIsPublicRoutableIP_ReturnsFalseForReservedIP(t *testing.T) {
	ip := net.ParseIP("192.0.2.1") // TEST-NET-1

	result := IsPublicRoutableIP(ip)

	assert.False(t, result)
}

func TestIsPublicRoutableIP_ReturnsFalseForMetadataIP(t *testing.T) {
	ip := net.ParseIP("169.254.169.254")

	result := IsPublicRoutableIP(ip)

	assert.False(t, result)
}

func TestIsReservedIP_ReturnsTrueForReservedRanges(t *testing.T) {
	testCases := []string{
		"0.0.0.1",        // "This" network
		"100.64.0.1",     // Shared address space
		"192.0.0.1",      // IETF protocol assignments
		"192.0.2.1",      // TEST-NET-1
		"198.18.0.1",     // Benchmarking
		"198.51.100.1",   // TEST-NET-2
		"203.0.113.1",    // TEST-NET-3
		"240.0.0.1",      // Reserved for future use
		"::",             // IPv6 unspecified
		"2001:db8::1",    // IPv6 documentation
	}

	for _, ipStr := range testCases {
		ip := net.ParseIP(ipStr)
		assert.True(t, IsReservedIP(ip), "Expected %s to be reserved", ipStr)
	}
}

func TestIsReservedIP_ReturnsFalseForNonReservedIP(t *testing.T) {
	ip := net.ParseIP("8.8.8.8")

	result := IsReservedIP(ip)

	assert.False(t, result)
}

func TestIsMetadataIP_ReturnsTrueForAWSMetadata(t *testing.T) {
	ip := net.ParseIP("169.254.169.254")

	result := IsMetadataIP(ip)

	assert.True(t, result)
}

func TestIsMetadataIP_ReturnsTrueForIPv6LinkLocal(t *testing.T) {
	ip := net.ParseIP("fe80::1")

	result := IsMetadataIP(ip)

	assert.True(t, result)
}

func TestIsMetadataIP_ReturnsTrueForIPv6ULA(t *testing.T) {
	ip := net.ParseIP("fc00::1")

	result := IsMetadataIP(ip)

	assert.True(t, result)
}

func TestIsMetadataIP_ReturnsFalseForNonMetadataIP(t *testing.T) {
	ip := net.ParseIP("8.8.8.8")

	result := IsMetadataIP(ip)

	assert.False(t, result)
}
