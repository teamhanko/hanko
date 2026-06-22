package utils

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	pictureURLReasonEmpty         = "empty"
	pictureURLReasonTooLong       = "too_long"
	pictureURLReasonInvalidURI    = "invalid_uri"
	pictureURLReasonInvalidScheme = "invalid_scheme"
	pictureURLReasonMissingHost   = "missing_host"
	pictureURLReasonHasUserInfo   = "has_userinfo"
)

// PictureURLError is returned by ValidatePictureURL when the URL is invalid.
type PictureURLError struct {
	Reason string
}

func (e PictureURLError) Error() string {
	return fmt.Sprintf("invalid picture url: %s", e.Reason)
}

// ValidatePictureURL returns nil if valid, otherwise an error that includes a reason string.
func ValidatePictureURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return PictureURLError{Reason: pictureURLReasonEmpty}
	}
	if len(raw) > 2048 {
		return PictureURLError{Reason: pictureURLReasonTooLong}
	}

	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return PictureURLError{Reason: pictureURLReasonInvalidURI}
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return PictureURLError{Reason: pictureURLReasonInvalidScheme}
	}

	if u.Host == "" {
		return PictureURLError{Reason: pictureURLReasonMissingHost}
	}

	// Disallow credentials in URL (user:pass@host).
	if u.User != nil {
		return PictureURLError{Reason: pictureURLReasonHasUserInfo}
	}

	return nil
}
