package webhooks

import (
	"context"
	"fmt"
	"net"
	"time"
)

// ValidatedDialer is a custom dialer that enforces connections only to pre-validated IP addresses.
// This prevents DNS rebinding attacks by ensuring the HTTP client connects to the same IPs
// that were validated during the security check, rather than re-resolving DNS.
type ValidatedDialer struct {
	// validatedIPs contains the list of IP addresses that passed security validation
	validatedIPs []net.IP

	// originalHost is the hostname from the original URL (used for TLS SNI)
	originalHost string

	// baseDialer is the underlying dialer used for actual connections
	baseDialer *net.Dialer

	// currentIPIndex tracks which IP to try next (for round-robin with multiple IPs)
	currentIPIndex int
}

// NewValidatedDialer creates a new dialer that only connects to the specified validated IPs.
func NewValidatedDialer(validatedIPs []net.IP, originalHost string) *ValidatedDialer {
	return &ValidatedDialer{
		validatedIPs:   validatedIPs,
		originalHost:   originalHost,
		baseDialer:     &net.Dialer{Timeout: 30 * time.Second},
		currentIPIndex: 0,
	}
}

// DialContext implements a custom dialer that connects only to pre-validated IPs.
// It ignores the 'address' parameter from the HTTP client and uses the validated IPs instead.
// This prevents DNS rebinding by forcing connections to IPs validated at config/trigger time.
func (d *ValidatedDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if len(d.validatedIPs) == 0 {
		return nil, fmt.Errorf("no validated IPs available for connection")
	}

	// Extract the port from the address (HTTP client provides "host:port")
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("failed to extract port from address '%s': %w", address, err)
	}

	// Try each validated IP until one succeeds or all fail
	var lastErr error
	startIdx := d.currentIPIndex

	for i := 0; i < len(d.validatedIPs); i++ {
		// Round-robin through IPs
		idx := (startIdx + i) % len(d.validatedIPs)
		ip := d.validatedIPs[idx]

		// Construct the target address using the validated IP and the original port
		targetAddr := net.JoinHostPort(ip.String(), port)

		conn, err := d.baseDialer.DialContext(ctx, network, targetAddr)
		if err == nil {
			// Update current index for next call (simple load balancing)
			d.currentIPIndex = (idx + 1) % len(d.validatedIPs)
			return conn, nil
		}

		lastErr = err
	}

	// All IPs failed
	if lastErr != nil {
		return nil, fmt.Errorf("all validated IPs failed to connect: %w", lastErr)
	}

	return nil, fmt.Errorf("failed to connect to any validated IP")
}
