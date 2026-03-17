package validation

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeError_Disabled(t *testing.T) {
	originalErr := errors.New("detailed error message with sensitive info")
	result := SanitizeError(originalErr, false)

	assert.Equal(t, originalErr, result, "should return original error when sanitization is disabled")
}

func TestSanitizeError_Nil(t *testing.T) {
	result := SanitizeError(nil, true)
	assert.Nil(t, result, "should return nil when error is nil")
}

func TestSanitizeError_DNSResolutionErrors(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedMsg   string
		shouldContain bool
	}{
		{
			name:        "failed to resolve error",
			err:         fmt.Errorf("failed to resolve webhook callback host 'internal.corp': no such host"),
			expectedMsg: ErrGenericValidationFailed,
		},
		{
			name:        "did not resolve error",
			err:         fmt.Errorf("webhook callback host 'secret.local' did not resolve to any IP addresses"),
			expectedMsg: ErrGenericValidationFailed,
		},
		{
			name:        "no such host error",
			err:         fmt.Errorf("lookup failed: no such host"),
			expectedMsg: ErrGenericValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeError(tt.err, true)
			assert.Equal(t, tt.expectedMsg, result.Error())

			// Verify detailed error is preserved
			sanitized, ok := result.(*SanitizedError)
			require.True(t, ok)
			assert.Equal(t, tt.err, sanitized.DetailedError)
		})
	}
}

func TestSanitizeError_HostBlockingErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "blocked host",
			err:         fmt.Errorf("host 'evil.com' is blocked"),
			expectedMsg: ErrGenericHostNotAllowed,
		},
		{
			name:        "blocked domain",
			err:         fmt.Errorf("host 'api.evil.com' matches a blocked domain"),
			expectedMsg: ErrGenericHostNotAllowed,
		},
		{
			name:        "not allowed host",
			err:         fmt.Errorf("host 'internal.local' is not allowed"),
			expectedMsg: ErrGenericHostNotAllowed,
		},
		{
			name:        "domain not in allowlist",
			err:         fmt.Errorf("host 'test.internal.local' is not in the allowed host/domain list"),
			expectedMsg: ErrGenericHostNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeError(tt.err, true)
			assert.Equal(t, tt.expectedMsg, result.Error())
		})
	}
}

func TestSanitizeError_IPValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "blocked IP",
			err:         fmt.Errorf("IP '192.168.1.1' is blocked"),
			expectedMsg: ErrGenericIPNotAllowed,
		},
		{
			name:        "non-public IP",
			err:         fmt.Errorf("non-public IP '10.0.0.1' is not allowed in public_only mode"),
			expectedMsg: ErrGenericIPNotAllowed,
		},
		{
			name:        "reserved IP",
			err:         fmt.Errorf("reserved IP '0.0.0.0' is blocked"),
			expectedMsg: ErrGenericIPNotAllowed,
		},
		{
			name:        "private IP",
			err:         fmt.Errorf("private IP '172.16.0.1' is not allowed"),
			expectedMsg: ErrGenericIPNotAllowed,
		},
		{
			name:        "resolved IP not allowed",
			err:         fmt.Errorf("resolved IP '10.0.0.5' for host 'internal.local' is not allowed: non-public IP"),
			expectedMsg: ErrGenericHostNotAllowed, // "host" keyword takes precedence
		},
		{
			name:        "CIDR match",
			err:         fmt.Errorf("IP '192.168.1.100' matches a blocked CIDR"),
			expectedMsg: ErrGenericIPNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeError(tt.err, true)
			assert.Equal(t, tt.expectedMsg, result.Error())
		})
	}
}

func TestSanitizeError_MetadataEndpointErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "metadata IP blocked",
			err:         fmt.Errorf("metadata endpoint IP '169.254.169.254' is blocked"),
			expectedMsg: ErrGenericIPNotAllowed,
		},
		{
			name:        "metadata host blocked",
			err:         fmt.Errorf("metadata endpoint host 'metadata.google.internal' is blocked"),
			expectedMsg: ErrGenericHostNotAllowed, // "host" keyword takes precedence over "metadata"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeError(tt.err, true)
			assert.Equal(t, tt.expectedMsg, result.Error())
		})
	}
}

func TestSanitizeError_RedirectErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "redirect target rejected",
			err:         fmt.Errorf("redirect target rejected by outbound policy: host not allowed"),
			expectedMsg: ErrGenericHostNotAllowed, // "host" keyword takes precedence over "redirect"
		},
		{
			name:        "too many redirects",
			err:         fmt.Errorf("too many redirects"),
			expectedMsg: ErrGenericRedirectNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeError(tt.err, true)
			assert.Equal(t, tt.expectedMsg, result.Error())
		})
	}
}

func TestSanitizeError_GenericErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "invalid URL",
			err:         fmt.Errorf("invalid webhook callback URL: parse error"),
			expectedMsg: ErrGenericValidationFailed,
		},
		{
			name:        "unknown error",
			err:         fmt.Errorf("some unknown error"),
			expectedMsg: ErrGenericValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeError(tt.err, true)
			assert.Equal(t, tt.expectedMsg, result.Error())
		})
	}
}

func TestSanitizeError_AlreadySanitized(t *testing.T) {
	originalErr := errors.New("detailed error")
	sanitized := &SanitizedError{
		DetailedError: originalErr,
		SanitizedMsg:  ErrGenericIPNotAllowed,
	}

	result := SanitizeError(sanitized, true)
	assert.Equal(t, sanitized, result, "should not double-sanitize")
}

func TestGetDetailedError(t *testing.T) {
	t.Run("sanitized error", func(t *testing.T) {
		originalErr := errors.New("detailed error with sensitive info")
		sanitized := SanitizeError(originalErr, true)

		detailed := GetDetailedError(sanitized)
		assert.Equal(t, originalErr, detailed)
	})

	t.Run("regular error", func(t *testing.T) {
		regularErr := errors.New("regular error")
		detailed := GetDetailedError(regularErr)
		assert.Equal(t, regularErr, detailed)
	})

	t.Run("nil error", func(t *testing.T) {
		detailed := GetDetailedError(nil)
		assert.Nil(t, detailed)
	})
}

func TestSanitizedError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	sanitized := &SanitizedError{
		DetailedError: originalErr,
		SanitizedMsg:  ErrGenericIPNotAllowed,
	}

	unwrapped := errors.Unwrap(sanitized)
	assert.Equal(t, originalErr, unwrapped)
}

func TestSanitizedError_Detailed(t *testing.T) {
	originalErr := errors.New("original detailed error")
	sanitized := &SanitizedError{
		DetailedError: originalErr,
		SanitizedMsg:  ErrGenericHostNotAllowed,
	}

	detailed := sanitized.Detailed()
	assert.Equal(t, originalErr, detailed)
}

func TestWrapWithContext(t *testing.T) {
	t.Run("with sanitization enabled", func(t *testing.T) {
		originalErr := fmt.Errorf("IP '10.0.0.1' is blocked")
		wrapped := WrapWithContext(originalErr, "validation failed", true)

		assert.NotNil(t, wrapped)
		assert.Equal(t, ErrGenericIPNotAllowed, wrapped.Error())

		detailed := GetDetailedError(wrapped)
		assert.Contains(t, detailed.Error(), "validation failed")
		assert.Contains(t, detailed.Error(), "IP '10.0.0.1' is blocked")
	})

	t.Run("with sanitization disabled", func(t *testing.T) {
		originalErr := fmt.Errorf("IP '10.0.0.1' is blocked")
		wrapped := WrapWithContext(originalErr, "validation failed", false)

		assert.Contains(t, wrapped.Error(), "validation failed")
		assert.Contains(t, wrapped.Error(), "IP '10.0.0.1' is blocked")
	})

	t.Run("nil error", func(t *testing.T) {
		wrapped := WrapWithContext(nil, "context", true)
		assert.Nil(t, wrapped)
	})
}

func TestSanitizeError_PreservesErrorChain(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)
	sanitizedErr := SanitizeError(wrappedErr, true)

	sanitized, ok := sanitizedErr.(*SanitizedError)
	require.True(t, ok)

	// Detailed error should preserve the chain
	assert.True(t, errors.Is(sanitized.DetailedError, baseErr))
}
