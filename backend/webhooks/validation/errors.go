package validation

import (
	"errors"
	"fmt"
	"strings"
)

// Sanitized error messages that don't leak internal network information
const (
	ErrGenericValidationFailed   = "callback URL validation failed"
	ErrGenericHostNotAllowed     = "callback URL not allowed"
	ErrGenericIPNotAllowed       = "callback destination not allowed"
	ErrGenericRedirectNotAllowed = "redirect destination not allowed"
)

// SanitizedError wraps an error with both detailed and sanitized messages
type SanitizedError struct {
	DetailedError error
	SanitizedMsg  string
}

func (e *SanitizedError) Error() string {
	return e.SanitizedMsg
}

func (e *SanitizedError) Unwrap() error {
	return e.DetailedError
}

// Detailed returns the original detailed error for logging
func (e *SanitizedError) Detailed() error {
	return e.DetailedError
}

// SanitizeError wraps an error with a sanitized message based on error type.
// Returns the original error if sanitization is disabled.
func SanitizeError(err error, sanitize bool) error {
	if err == nil {
		return nil
	}

	if !sanitize {
		return err
	}

	// Already sanitized
	if _, ok := err.(*SanitizedError); ok {
		return err
	}

	errMsg := err.Error()
	sanitizedMsg := categorizeError(errMsg)

	return &SanitizedError{
		DetailedError: err,
		SanitizedMsg:  sanitizedMsg,
	}
}

// categorizeError maps detailed error messages to generic sanitized versions
func categorizeError(errMsg string) string {
	errLower := strings.ToLower(errMsg)

	// DNS resolution errors
	if strings.Contains(errLower, "failed to resolve") ||
		strings.Contains(errLower, "did not resolve") ||
		strings.Contains(errLower, "no such host") {
		return ErrGenericValidationFailed
	}

	// Host/domain blocking errors
	if strings.Contains(errLower, "host") && (strings.Contains(errLower, "blocked") || strings.Contains(errLower, "not allowed")) {
		return ErrGenericHostNotAllowed
	}

	if strings.Contains(errLower, "domain") && (strings.Contains(errLower, "blocked") || strings.Contains(errLower, "not allowed")) {
		return ErrGenericHostNotAllowed
	}

	// IP validation errors
	if strings.Contains(errLower, "ip") && (strings.Contains(errLower, "blocked") || strings.Contains(errLower, "not allowed")) {
		return ErrGenericIPNotAllowed
	}

	if strings.Contains(errLower, "resolved ip") {
		return ErrGenericIPNotAllowed
	}

	// Metadata endpoint errors
	if strings.Contains(errLower, "metadata") {
		return ErrGenericIPNotAllowed
	}

	// Reserved/private IP errors
	if strings.Contains(errLower, "non-public") ||
		strings.Contains(errLower, "reserved") ||
		strings.Contains(errLower, "private") {
		return ErrGenericIPNotAllowed
	}

	// CIDR matching errors
	if strings.Contains(errLower, "cidr") {
		return ErrGenericIPNotAllowed
	}

	// Redirect errors
	if strings.Contains(errLower, "redirect") {
		return ErrGenericRedirectNotAllowed
	}

	// Allowlist errors
	if strings.Contains(errLower, "not in the allowed") ||
		strings.Contains(errLower, "not allowed unless explicitly allowlisted") {
		return ErrGenericHostNotAllowed
	}

	// Default to generic validation failure
	return ErrGenericValidationFailed
}

// GetDetailedError extracts the detailed error from a SanitizedError for logging.
// Returns the original error if it's not a SanitizedError.
func GetDetailedError(err error) error {
	if err == nil {
		return nil
	}

	var sanitizedErr *SanitizedError
	if errors.As(err, &sanitizedErr) {
		return sanitizedErr.DetailedError
	}

	return err
}

// WrapWithContext adds context to an error and applies sanitization if enabled
func WrapWithContext(err error, context string, sanitize bool) error {
	if err == nil {
		return nil
	}

	// Get the detailed error for wrapping
	detailedErr := GetDetailedError(err)
	wrappedErr := fmt.Errorf("%s: %w", context, detailedErr)

	return SanitizeError(wrappedErr, sanitize)
}
