package utils

import (
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

// ValidatePictureURL returns "" if valid, otherwise a stable reason code.
func ValidatePictureURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pictureURLReasonEmpty
	}
	if len(raw) > 2048 {
		return pictureURLReasonTooLong
	}

	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return pictureURLReasonInvalidURI
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return pictureURLReasonInvalidScheme
	}

	if u.Host == "" {
		return pictureURLReasonMissingHost
	}

	// Disallow credentials in URL (user:pass@host).
	if u.User != nil {
		return pictureURLReasonHasUserInfo
	}

	return ""
}
