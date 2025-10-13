package device_trust

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flow_api/services"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
)

type ScheduleTrustDeviceState struct {
	shared.Action
}

func (h ScheduleTrustDeviceState) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	if !deps.Cfg.MFA.Enabled || deps.Cfg.MFA.DeviceTrustPolicy != "prompt" {
		return nil
	}

	if c.IsFlow(shared.FlowLogin) && c.Stash().Get(shared.StashPathLoginMethod).String() == "passkey" {
		return nil
	}

	if !c.Stash().Get(shared.StashPathUserHasSecurityKey).Bool() &&
		!c.Stash().Get(shared.StashPathUserHasOTPSecret).Bool() {
		return nil
	}

	deviceTrustService := services.DeviceTrustService{
		Persister:   deps.Persister.GetTrustedDevicePersisterWithConnection(deps.Tx),
		Cfg:         deps.Cfg,
		HttpContext: deps.HttpContext,
	}

	userID := uuid.FromStringOrNil(c.Stash().Get(shared.StashPathUserID).String())

	if !deviceTrustService.CheckDeviceTrust(userID) {
		c.ScheduleStates(shared.StateDeviceTrust)
	}

	return nil
}
