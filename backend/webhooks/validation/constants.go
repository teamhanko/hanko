package validation

// ReservedIPCIDRs contains reserved and special-use IPv4 and IPv6 CIDR ranges that should be blocked
// for outbound webhook callbacks to prevent SSRF attacks.
//
// References:
// - RFC 5735 (IPv4 Special-Use Addresses)
// - RFC 6890 (Special-Purpose IP Address Registries)
// - RFC 4193 (IPv6 Unique Local Addresses)
// - RFC 4291 (IPv6 Address Architecture)
var ReservedIPCIDRs = []string{
	// IPv4 reserved ranges
	"0.0.0.0/8",       // "This" network (RFC 5735)
	"100.64.0.0/10",   // Shared address space (RFC 6598)
	"192.0.0.0/24",    // IETF protocol assignments (RFC 5736)
	"192.0.2.0/24",    // TEST-NET-1 (RFC 5737)
	"198.18.0.0/15",   // Benchmarking (RFC 2544)
	"198.51.100.0/24", // TEST-NET-2 (RFC 5737)
	"203.0.113.0/24",  // TEST-NET-3 (RFC 5737)
	"240.0.0.0/4",     // Reserved for future use (RFC 1112)

	// IPv6 reserved ranges
	"::/128",        // Unspecified address
	"100::/64",      // Discard prefix (RFC 6666)
	"2001:db8::/32", // Documentation prefix (RFC 3849)
}

// MetadataEndpointCIDRs cntains cloud provider metadata service endpoints that should be blocked
// to prevent credential theft and information disclosure.
//
// References:
// - AWS: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html
// - GCP: https://cloud.google.com/compute/docs/metadata/overview
// - Azure: https://learn.microsoft.com/en-us/azure/virtual-machines/instance-metadata-service
// - Oracle Cloud: https://docs.oracle.com/en-us/iaas/Content/Compute/Tasks/gettingmetadata.htm
var MetadataEndpointCIDRs = []string{
	// Common IPv4 provider range, see references
	"169.254.169.254/32",

	// IPv6 link-local addresses (used by various cloud providers for metadata)
	// This is a broad range but necessary to block metadata services
	"fe80::/10",

	// IPv6 Unique Local Addresses (private IPv6, similar to RFC 1918)
	"fc00::/7",

	// Common IPv6 provider range, see references
	"fd00:ec2::254/128",
}

// AdditionalMetadataHosts contains hostname-based metadata endpoints
// that should be blocked in addition to IP-based blocking.
var AdditionalMetadataHosts = []string{
	"metadata.google.internal",     // GCP metadata service
	"metadata.goog",                // GCP metadata service (alternative)
	"169.254.169.254.nip.io",       // Common bypass attempt
	"169.254.169.254.xip.io",       // Common bypass attempt
	"169.254.169.254.sslip.io",     // Common bypass attempt
	"metadata",                     // Generic metadata hostname
	"instance-data",                // AWS instance data
	"169-254-169-254.ec2.internal", // AWS internal hostname format
}
