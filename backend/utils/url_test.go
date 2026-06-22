package utils

import (
	"strings"
	"testing"
)

func TestValidatePictureURL(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		wantReason string
	}{
		{
			name:       "valid https URL",
			raw:        "https://example.com/image.png",
			wantReason: "",
		},
		{
			name:       "valid http URL",
			raw:        "http://example.com/image.png",
			wantReason: "",
		},
		{
			name:       "valid URL with query",
			raw:        "https://example.com/image.png?size=large",
			wantReason: "",
		},
		{
			name:       "empty string",
			raw:        "",
			wantReason: pictureURLReasonEmpty,
		},
		{
			name:       "only whitespace",
			raw:        "   ",
			wantReason: pictureURLReasonEmpty,
		},
		{
			name:       "too long URL",
			raw:        "https://example.com/" + strings.Repeat("a", 2048),
			wantReason: pictureURLReasonTooLong,
		},
		{
			name:       "invalid URI",
			raw:        "://example.com",
			wantReason: pictureURLReasonInvalidURI,
		},
		{
			name:       "invalid scheme (ftp)",
			raw:        "ftp://example.com/image.png",
			wantReason: pictureURLReasonInvalidScheme,
		},
		{
			name:       "invalid scheme (data)",
			raw:        "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
			wantReason: pictureURLReasonInvalidScheme,
		},
		{
			name:       "missing host",
			raw:        "https:///path",
			wantReason: pictureURLReasonMissingHost,
		},
		{
			name:       "has user info",
			raw:        "https://user:pass@example.com/image.png",
			wantReason: pictureURLReasonHasUserInfo,
		},
		{
			name:       "valid URL with spaces",
			raw:        "  https://example.com/image.png  ",
			wantReason: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePictureURL(tt.raw)

			if tt.wantReason == "" {
				if err != nil {
					t.Fatalf("ValidatePictureURL(%q) error = %v, want nil", tt.raw, err)
				}
				return
			}

			if err == nil {
				t.Fatalf("ValidatePictureURL(%q) error = nil, want reason %q", tt.raw, tt.wantReason)
			}

			perr, ok := err.(PictureURLError)
			if !ok {
				t.Fatalf("ValidatePictureURL(%q) error type = %T, want utils.PictureURLError", tt.raw, err)
			}

			if perr.Reason != tt.wantReason {
				t.Fatalf("ValidatePictureURL(%q) reason = %q, want %q", tt.raw, perr.Reason, tt.wantReason)
			}
		})
	}
}
