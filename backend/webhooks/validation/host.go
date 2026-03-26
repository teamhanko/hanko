package validation

import (
	"strings"
)

// NormalizeHost normalizes a hostname by converting to lowercase,
// trimming whitespace, and removing trailing dots.
func NormalizeHost(value string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
}

// MatchesHost checks if the given host exactly matches any host in the provided list.
// Comparison is case-insensitive and normalizes both the input and list values.
func MatchesHost(host string, values []string) bool {
	host = NormalizeHost(host)
	for _, value := range values {
		if host == NormalizeHost(value) {
			return true
		}
	}
	return false
}

// MatchesDomain checks if the given host matches any domain in the provided list.
// It matches both exact domain matches and subdomain matches.
// For example, "api.example.com" matches domain "example.com", but "badexample.com" does not.
func MatchesDomain(host string, domains []string) bool {
	host = NormalizeHost(host)
	for _, domain := range domains {
		normalized := NormalizeHost(domain)
		if host == normalized || strings.HasSuffix(host, "."+normalized) {
			return true
		}
	}
	return false
}

// IsMetadataHost checks if the given hostname is a known metadata service endpoint.
// This provides an additional layer of protection beyond IP-based blocking.
func IsMetadataHost(host string) bool {
	return MatchesHost(host, AdditionalMetadataHosts)
}
