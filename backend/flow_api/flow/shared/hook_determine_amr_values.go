package shared

import (
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
)

type DetermineAMRValues struct {
	Action
}

func (h DetermineAMRValues) Execute(c flowpilot.HookExecutionContext) error {
	loginMethod := c.Stash().Get(StashPathLoginMethod).String()
	mfaMethod := c.Stash().Get(StashPathMFAUsageMethod).String()
	thirdPartyProvider := c.Stash().Get(StashPathThirdPartyProvider).String()

	var amr []string

	switch loginMethod {
	case "password":
		amr = append(amr, "pwd")
	case "passkey":
		amr = append(amr, "passkey")
	case "passcode":
		amr = append(amr, "otp")
	case "third_party":
		if thirdPartyProvider != "" {
			amr = append(amr, "ext:"+thirdPartyProvider)
		} else {
			amr = append(amr, "ext")
		}
	}

	switch mfaMethod {
	case "totp":
		amr = append(amr, "totp")
	case "security_key":
		amr = append(amr, "security_key")
	}

	if len(amr) == 0 {
		return nil
	}

	return c.Stash().Set(StashPathAMRValues, amr)
}
