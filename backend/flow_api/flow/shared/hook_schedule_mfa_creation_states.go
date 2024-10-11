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
	mfaAcquireConfigured := mfaConfig.Enabled &&
		(mfaConfig.SecurityKeys.Enabled || mfaConfig.TOTP.Enabled) &&
		((c.GetFlowName() == FlowLogin && mfaConfig.AcquireOnLogin) ||
			(c.GetFlowName() == FlowRegistration && mfaConfig.AcquireOnRegistration))
	userHasSecurityKey := c.Stash().Get(StashPathUserHasSecurityKey).Bool()
	attachmentSupported := c.Stash().Get(StashPathSecurityKeyAttachmentSupported).Bool()
	userHasOTPSecret := c.Stash().Get(StashPathUserHasOTPSecret).Bool()

	if !userHasSecurityKey && !userHasOTPSecret && mfaLoginEnabled && mfaAcquireConfigured {
		// The user has no MFA methods set up but MFA is enabled and configured for acquisition.

		deviceSupportsMFAMethod := mfaConfig.TOTP.Enabled || attachmentSupported

		if !deviceSupportsMFAMethod {
			// The device or browser does not support a suitable MFA method.

			if !mfaConfig.Optional {
				// Show error when onboarding is required.
				c.SetFlowError(ErrorPlatformAuthenticatorRequired)
			} else {
				// Skip onboarding, when optional.
			}
		} else {
			c.ScheduleStates(StateMFAMethodChooser)
		}
	}

	return nil
}
