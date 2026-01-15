package device_trust

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flow_api/services"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
	"net/http"
)

type IssueTrustDeviceCookie struct {
	shared.Action
}

func (h IssueTrustDeviceCookie) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	if deps.Cfg.MFA.DeviceTrustPolicy == "never" ||
		(deps.Cfg.MFA.DeviceTrustPolicy == "prompt" && !c.Stash().Get(shared.StashPathDeviceTrustGranted).Bool()) {
		return nil
	}

	if !c.Stash().Get(shared.StashPathUserID).Exists() {
		return fmt.Errorf("user id does not exist in the stash")
	}

	userID, err := uuid.FromString(c.Stash().Get(shared.StashPathUserID).String())
	if err != nil {
		return fmt.Errorf("failed to parse stashed user_id into a uuid: %w", err)
	}

	deviceTrustService := services.DeviceTrustService{
		Persister:   deps.Persister.GetTrustedDevicePersisterWithConnection(deps.Tx),
		Cfg:         deps.Cfg,
		HttpContext: deps.HttpContext,
	}

	// Generate new token for this user
	deviceToken, err := deviceTrustService.GenerateRandomToken(64)
	if err != nil {
		return fmt.Errorf("failed to generate trusted device token: %w", err)
	}

	name := deps.Cfg.MFA.DeviceTrustCookieName
	maxAge := int(deps.Cfg.MFA.DeviceTrustDuration.Seconds())

	if maxAge > 0 {
		err = deviceTrustService.CreateTrustedDevice(userID, deviceToken)
		if err != nil {
			return fmt.Errorf("failed to store trusted device: %w", err)
		}
	}

	// Read existing cookie entries for multi-user support
	var entries []services.DeviceTrustEntry
	existingCookie, _ := deps.HttpContext.Cookie(name)
	if existingCookie != nil {
		entries = deviceTrustService.ParseDeviceTrustCookie(existingCookie.Value)
	}

	// Remove existing entry for this user (if any)
	var filteredEntries []services.DeviceTrustEntry
	for _, entry := range entries {
		if entry.UserID.String() != userID.String() {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	// Add new entry at the front (most recent)
	newEntry := services.DeviceTrustEntry{
		UserID:      userID,
		DeviceToken: deviceToken,
	}
	entries = append([]services.DeviceTrustEntry{newEntry}, filteredEntries...)

	// Enforce max users limit
	maxUsers := deps.Cfg.MFA.DeviceTrustMaxUsersPerDevice
	if maxUsers <= 0 {
		maxUsers = 20 // Default
	}
	if len(entries) > maxUsers {
		entries = entries[:maxUsers]
	}

	// Serialize composite cookie value
	cookieValue := deviceTrustService.SerializeDeviceTrustCookie(entries)

	cookie := new(http.Cookie)
	cookie.Name = name
	cookie.Value = cookieValue
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.MaxAge = maxAge
	cookie.SameSite = http.SameSiteNoneMode

	deps.HttpContext.SetCookie(cookie)

	return nil
}
