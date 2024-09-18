package registration

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ScheduleMFACreationStates struct {
	shared.Action
}

func (h ScheduleMFACreationStates) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	userHasWebauthn := c.Stash().Get(shared.StashPathUserHasWebauthnCredential).Bool()
	mfaConfig := deps.Cfg.MFA

	if !userHasWebauthn && mfaConfig.Enabled && mfaConfig.AcquireOnRegistration &&
		(mfaConfig.SecurityKeys.Enabled || mfaConfig.TOTP.Enabled) {
		c.ScheduleStates(shared.StateMFAMethodChooser)
	}

	return nil
}
