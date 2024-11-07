package device_trust

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"net/http"
)

type IssueTrustDeviceCookie struct {
	shared.Action
}

func (h IssueTrustDeviceCookie) Execute(c flowpilot.HookExecutionContext) error {
	var err error

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

	deviceToken, err := deviceTrustService.GenerateRandomToken(62)
	if err != nil {
		return fmt.Errorf("failed to generate trusted device token: %w", err)
	}

	name := deps.Cfg.MFA.DeviceTrustCookieName
	maxAge := int(deps.Cfg.MFA.DeviceTrustDuration.Seconds())

	if maxAge > 0 {
		err = deviceTrustService.CreateTrustedDevice(userID, deviceToken)
		if err != nil {
			return fmt.Errorf("failed to storer trusted device: %w", err)
		}
	}

	cookie := new(http.Cookie)
	cookie.Name = name
	cookie.Value = deviceToken
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.MaxAge = maxAge

	deps.HttpContext.SetCookie(cookie)

	return nil
}
