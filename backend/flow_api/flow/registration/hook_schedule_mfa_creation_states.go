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
	mfaConfig := deps.Cfg.MFA

	if !mfaConfig.Enabled {
		return nil
	}

	passcodeLoginEligible := c.Stash().Get(shared.StashPathEmail).Exists() && deps.Cfg.Email.UseForAuthentication
	useHasPassword := c.Stash().Get(shared.StashPathUserHasPassword).Bool()

	if (useHasPassword || passcodeLoginEligible) &&
		mfaConfig.AcquireOnRegistration &&
		(mfaConfig.SecurityKeys.Enabled || mfaConfig.TOTP.Enabled) {
		c.ScheduleStates(shared.StateMFAMethodChooser)
	}

	return nil
}
