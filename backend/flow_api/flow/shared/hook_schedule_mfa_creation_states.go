package shared

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ScheduleMFACreationStates struct {
	Action
}

func (h ScheduleMFACreationStates) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	if c.IsStateScheduled(StatePasswordCreation) ||
		c.Stash().Get(StashPathPasswordRecoveryPending).Bool() {
		// Delay MFA onboarding until a password has eventually been set or updated.
		return nil
	}

	if c.StateVisited(StateMFAMethodChooser) {
		// Show MFA onboarding only once within a flow unless states have been reverted.
		return nil
	}

	mfaConfig := deps.Cfg.MFA
	passwordsEnabled := deps.Cfg.Password.Enabled
	passcodeEmailsEnabled := deps.Cfg.Email.Enabled && deps.Cfg.Email.UseForAuthentication
	userHasEmail := c.Stash().Get(StashPathEmail).Exists() || c.Stash().Get(StashPathUserHasEmails).Bool()
	userHasPassword := c.Stash().Get(StashPathUserHasPassword).Bool()
	mfaLoginEnabled := (passwordsEnabled && userHasPassword) || (passcodeEmailsEnabled && userHasEmail)
	mfaMethodsEnabled := mfaConfig.SecurityKeys.Enabled || mfaConfig.TOTP.Enabled
	acquireMFAMethod := (c.GetFlowName() == FlowLogin && mfaConfig.AcquireOnLogin) ||
		(c.GetFlowName() == FlowRegistration && mfaConfig.AcquireOnRegistration)
	userHasSecurityKey := c.Stash().Get(StashPathUserHasSecurityKey).Bool()
	userHasOTPSecret := c.Stash().Get(StashPathUserHasOTPSecret).Bool()

	if !userHasSecurityKey && !userHasOTPSecret &&
		mfaConfig.Enabled && mfaLoginEnabled &&
		acquireMFAMethod && mfaMethodsEnabled {
		c.ScheduleStates(StateMFAMethodChooser)
	}

	return nil
}
