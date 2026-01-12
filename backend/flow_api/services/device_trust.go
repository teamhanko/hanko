package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"time"
)

// DeviceTrustEntry represents a single user's device trust token entry
type DeviceTrustEntry struct {
	UserID      uuid.UUID
	DeviceToken string
}

const (
	// entrySeparator separates multiple user entries in the cookie
	entrySeparator = "|"
	// fieldSeparator separates user ID from token within an entry
	fieldSeparator = ":"
)

type DeviceTrustService struct {
	Persister   persistence.TrustedDevicePersister
	Cfg         config.Config
	HttpContext echo.Context
}

func (s DeviceTrustService) CreateTrustedDevice(userID uuid.UUID, deviceToken string) error {
	deviceID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to generate device id: %w", err)
	}

	trustedDeviceModel := models.TrustedDevice{
		ID:          deviceID,
		UserID:      userID,
		DeviceToken: deviceToken,
		ExpiresAt:   time.Now().Add(s.Cfg.MFA.DeviceTrustDuration).UTC(),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	err = s.Persister.Create(trustedDeviceModel)
	if err != nil {
		return fmt.Errorf("failed to store trusted device: %w", err)
	}

	return nil
}

func (s DeviceTrustService) CheckDeviceTrust(userID uuid.UUID) bool {
	if userID.IsNil() || s.Cfg.MFA.DeviceTrustPolicy == "never" {
		return false
	}

	cookieName := s.Cfg.MFA.DeviceTrustCookieName
	cookie, _ := s.HttpContext.Cookie(cookieName)

	if cookie == nil {
		return false
	}

	entries := s.ParseDeviceTrustCookie(cookie.Value)

	// Handle legacy format (single token without user ID)
	if entries == nil && cookie.Value != "" {
		// Legacy: look up token in DB to check if it belongs to this user
		trustedDevice, err := s.Persister.FindByDeviceToken(cookie.Value)
		if err == nil && trustedDevice != nil &&
			time.Now().UTC().Before(trustedDevice.ExpiresAt.UTC()) &&
			trustedDevice.UserID.String() == userID.String() {
			return true
		}
		return false
	}

	// New format: find entry for this user
	for _, entry := range entries {
		if entry.UserID.String() == userID.String() {
			trustedDevice, err := s.Persister.FindByDeviceToken(entry.DeviceToken)
			if err == nil && trustedDevice != nil &&
				time.Now().UTC().Before(trustedDevice.ExpiresAt.UTC()) &&
				trustedDevice.UserID.String() == userID.String() {
				return true
			}
		}
	}

	return false
}

func (s DeviceTrustService) GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ParseDeviceTrustCookie parses a composite device trust cookie value into individual entries.
// Returns nil if the cookie is empty or in legacy format (single token without user ID).
// Legacy format detection: no separators means it's a single token.
func (s DeviceTrustService) ParseDeviceTrustCookie(cookieValue string) []DeviceTrustEntry {
	if cookieValue == "" {
		return nil
	}

	// Legacy format detection (no separators = single token)
	if !strings.Contains(cookieValue, entrySeparator) && !strings.Contains(cookieValue, fieldSeparator) {
		return nil // Caller handles legacy migration
	}

	var entries []DeviceTrustEntry
	parts := strings.Split(cookieValue, entrySeparator)

	for _, part := range parts {
		fields := strings.SplitN(part, fieldSeparator, 2)
		if len(fields) != 2 {
			continue // Skip malformed entries
		}

		userID, err := uuid.FromString(fields[0])
		if err != nil {
			continue // Skip invalid user IDs
		}

		entries = append(entries, DeviceTrustEntry{
			UserID:      userID,
			DeviceToken: fields[1],
		})
	}

	return entries
}

// SerializeDeviceTrustCookie serializes device trust entries into a composite cookie value.
// Format: <user_id_1>:<token_1>|<user_id_2>:<token_2>|...
func (s DeviceTrustService) SerializeDeviceTrustCookie(entries []DeviceTrustEntry) string {
	if len(entries) == 0 {
		return ""
	}

	parts := make([]string, len(entries))
	for i, entry := range entries {
		parts[i] = entry.UserID.String() + fieldSeparator + entry.DeviceToken
	}

	return strings.Join(parts, entrySeparator)
}
