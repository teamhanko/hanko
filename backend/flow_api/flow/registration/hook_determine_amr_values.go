package registration

import (
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
)

type DetermineAMRValues struct {
	shared.Action
}

func (h DetermineAMRValues) Execute(c flowpilot.HookExecutionContext) error {
	if !c.IsFlow(shared.FlowRegistration) {
		return nil
	}

	var amr []string

	if c.Stash().Get(shared.StashPathRegistrationAMRUsedPasscode).Bool() {
		amr = append(amr, "otp")
	}

	if c.Stash().Get(shared.StashPathRegistrationAMRUsedThirdParty).Bool() {
		provider := c.Stash().Get(shared.StashPathRegistrationAMRUsedThirdPartyProvider).String()
		if provider != "" {
			amr = append(amr, "ext:"+provider)
		} else {
			amr = append(amr, "ext")
		}
	}

	if c.Stash().Get(shared.StashPathRegistrationAMREnrolledPwd).Bool() {
		amr = append(amr, "pwd")
	}
	if c.Stash().Get(shared.StashPathRegistrationAMREnrolledPasskey).Bool() {
		amr = append(amr, "passkey")
	}
	if c.Stash().Get(shared.StashPathRegistrationAMREnrolledTotp).Bool() {
		amr = append(amr, "totp")
	}
	if c.Stash().Get(shared.StashPathRegistrationAMREnrolledSecurityKey).Bool() {
		amr = append(amr, "security_key")
	}

	if len(amr) == 0 {
		return nil
	}

	return c.Stash().Set(shared.StashPathAMRValues, amr)
}
