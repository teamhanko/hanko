package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
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
	if !userID.IsNil() && s.Cfg.MFA.DeviceTrustPolicy != "never" {
		cookieName := s.Cfg.MFA.DeviceTrustCookieName
		cookie, _ := s.HttpContext.Cookie(cookieName)

		if cookie != nil {
			deviceToken := cookie.Value
			trustedDeviceModel, err := s.Persister.FindByDeviceToken(deviceToken)

			if err == nil && trustedDeviceModel != nil &&
				time.Now().UTC().Before(trustedDeviceModel.ExpiresAt.UTC()) &&
				trustedDeviceModel.UserID.String() == userID.String() {
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
