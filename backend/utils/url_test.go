package utils

import (
	"strings"
	"testing"
)

func TestValidatePictureURL(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected string
	}{
		{
			name:     "valid https URL",
			raw:      "https://example.com/image.png",
			expected: "",
		},
		{
			name:     "valid http URL",
			raw:      "http://example.com/image.png",
			expected: "",
		},
		{
			name:     "valid URL with query",
			raw:      "https://example.com/image.png?size=large",
			expected: "",
		},
		{
			name:     "empty string",
			raw:      "",
			expected: pictureURLReasonEmpty,
		},
		{
			name:     "only whitespace",
			raw:      "   ",
			expected: pictureURLReasonEmpty,
		},
		{
			name:     "too long URL",
			raw:      "https://example.com/" + strings.Repeat("a", 2048),
			expected: pictureURLReasonTooLong,
		},
		{
			name:     "invalid URI",
			raw:      "://example.com",
			expected: pictureURLReasonInvalidURI,
		},
		{
			name:     "invalid scheme (ftp)",
			raw:      "ftp://example.com/image.png",
			expected: pictureURLReasonInvalidScheme,
		},
		{
			name:     "invalid scheme (data)",
			raw:      "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
			expected: pictureURLReasonInvalidScheme,
		},
		{
			name:     "missing host",
			raw:      "https:///path",
			expected: pictureURLReasonMissingHost,
		},
		{
			name:     "has user info",
			raw:      "https://user:pass@example.com/image.png",
			expected: pictureURLReasonHasUserInfo,
		},
		{
			name:     "valid URL with spaces",
			raw:      "  https://example.com/image.png  ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePictureURL(tt.raw)
			if got != tt.expected {
				t.Errorf("ValidatePictureURL(%q) = %q, want %q", tt.raw, got, tt.expected)
			}
		})
	}
}
