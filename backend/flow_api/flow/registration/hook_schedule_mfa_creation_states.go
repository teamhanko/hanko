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
	passwordsEnabled := deps.Cfg.Password.Enabled
	passcodeEmailsEnabled := deps.Cfg.Email.Enabled && deps.Cfg.Email.UseForAuthentication
	userHasEmail := c.Stash().Get(shared.StashPathEmail).Exists() || c.Stash().Get(shared.StashPathUserHasEmails).Bool()
	userHasPassword := c.Stash().Get(shared.StashPathUserHasPassword).Bool()
	mfaLoginEnabled := (passwordsEnabled && userHasPassword) || (passcodeEmailsEnabled && userHasEmail)
	mfaMethodsEnabled := mfaConfig.SecurityKeys.Enabled || mfaConfig.TOTP.Enabled
	acquireMFAMethod := (c.GetFlowName() == shared.FlowLogin && mfaConfig.AcquireOnLogin) ||
		(c.GetFlowName() == shared.FlowRegistration && mfaConfig.AcquireOnRegistration)

	if mfaConfig.Enabled && mfaLoginEnabled && acquireMFAMethod && mfaMethodsEnabled {
		c.ScheduleStates(shared.StateMFAMethodChooser)
	}

	return nil
}
