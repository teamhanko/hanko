package validation

import (
	"net"
	"strings"
)

// IPMatchesCIDRs checks if the given IP address matches any of the provided CIDR ranges.
// Invalid CIDR ranges in the list are silently skipped.
func IPMatchesCIDRs(ip net.IP, cidrs []string) bool {
	for _, value := range cidrs {
		_, network, err := net.ParseCIDR(strings.TrimSpace(value))
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// IsPublicRoutableIP returns true if the IP address is a publicly routable IP.
// It returns false for loopback, private, multicast, unspecified, link-local,
// reserved, and metadata service IPs.
func IsPublicRoutableIP(ip net.IP) bool {
	if ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsMulticast() ||
		ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		IsReservedIP(ip) ||
		IsMetadataIP(ip) {
		return false
	}

	return true
}

// IsReservedIP checks if the IP address falls within reserved or special-use IP ranges.
// See ReservedIPCIDRs in constants.go for the full list of reserved ranges.
func IsReservedIP(ip net.IP) bool {
	return IPMatchesCIDRs(ip, ReservedIPCIDRs)
}

// IsMetadataIP checks if the IP address is a cloud provider metadata service endpoint.
// See MetadataEndpointCIDRs in constants.go for the full list of metadata endpoints.
func IsMetadataIP(ip net.IP) bool {
	return IPMatchesCIDRs(ip, MetadataEndpointCIDRs)
}
