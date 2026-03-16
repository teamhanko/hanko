package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeHost_NormalizesCorrectly(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{" Example.COM. ", "example.com"},
		{"EXAMPLE.COM", "example.com"},
		{"example.com.", "example.com"},
		{"  example.com  ", "example.com"},
		{"Example.Com", "example.com"},
	}

	for _, tc := range testCases {
		result := NormalizeHost(tc.input)
		assert.Equal(t, tc.expected, result, "Input: %s", tc.input)
	}
}

func TestMatchesHost_MatchesExactHost(t *testing.T) {
	result := MatchesHost("example.com", []string{"example.com"})

	assert.True(t, result)
}

func TestMatchesHost_MatchesCaseInsensitive(t *testing.T) {
	result := MatchesHost("Example.COM", []string{"example.com"})

	assert.True(t, result)
}

func TestMatchesHost_DoesNotMatchSubdomain(t *testing.T) {
	result := MatchesHost("api.example.com", []string{"example.com"})

	assert.False(t, result)
}

func TestMatchesHost_DoesNotMatchDifferentHost(t *testing.T) {
	result := MatchesHost("other.com", []string{"example.com"})

	assert.False(t, result)
}

func TestMatchesDomain_MatchesExactDomain(t *testing.T) {
	result := MatchesDomain("example.com", []string{"example.com"})

	assert.True(t, result)
}

func TestMatchesDomain_MatchesSubdomain(t *testing.T) {
	result := MatchesDomain("api.example.com", []string{"example.com"})

	assert.True(t, result)
}

func TestMatchesDomain_MatchesDeepSubdomain(t *testing.T) {
	result := MatchesDomain("a.b.c.example.com", []string{"example.com"})

	assert.True(t, result)
}

func TestMatchesDomain_DoesNotMatchDifferentDomain(t *testing.T) {
	result := MatchesDomain("badexample.com", []string{"example.com"})

	assert.False(t, result)
}

func TestMatchesDomain_DoesNotMatchPartialMatch(t *testing.T) {
	result := MatchesDomain("notexample.com", []string{"example.com"})

	assert.False(t, result)
}

func TestMatchesDomain_MatchesCaseInsensitive(t *testing.T) {
	result := MatchesDomain("API.Example.COM", []string{"example.com"})

	assert.True(t, result)
}

func TestIsMetadataHost_ReturnsTrueForKnownMetadataHosts(t *testing.T) {
	testCases := []string{
		"metadata.google.internal",
		"metadata.goog",
		"169.254.169.254.nip.io",
		"metadata",
		"instance-data",
	}

	for _, host := range testCases {
		result := IsMetadataHost(host)
		assert.True(t, result, "Expected %s to be a metadata host", host)
	}
}

func TestIsMetadataHost_ReturnsFalseForNonMetadataHost(t *testing.T) {
	result := IsMetadataHost("example.com")

	assert.False(t, result)
}

func TestIsMetadataHost_MatchesCaseInsensitive(t *testing.T) {
	result := IsMetadataHost("METADATA.GOOGLE.INTERNAL")

	assert.True(t, result)
}
