package login

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ScheduleOnboardingStates struct {
	shared.Action
}

func (h ScheduleOnboardingStates) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	userHasPassword := deps.Cfg.Password.Enabled && c.Stash().Get(shared.StashPathUserHasPassword).Bool()
	userHasPasskey := deps.Cfg.Passkey.Enabled && c.Stash().Get(shared.StashPathUserHasWebauthnCredential).Bool()
	userHasUsername := deps.Cfg.Username.Enabled && c.Stash().Get(shared.StashPathUserHasUsername).Bool()
	userHasEmail := deps.Cfg.Email.Enabled && c.Stash().Get(shared.StashPathUserHasEmails).Bool()

	if err := c.Stash().Set(shared.StashPathUserHasPassword, userHasPassword); err != nil {
		return fmt.Errorf("failed to set user_has_password to the stash: %w", err)
	}

	if err := c.Stash().Set(shared.StashPathUserHasWebauthnCredential, userHasPasskey); err != nil {
		return fmt.Errorf("failed to set user_has_webauthn_credential to the stash: %w", err)
	}

	userDetailOnboardingStates := h.determineUserDetailOnboardingStates(c, userHasUsername, userHasEmail)
	credentialOnboardingStates := h.determineCredentialOnboardingStates(c, userHasPasskey, userHasPassword)

	return c.ScheduleStates(append(userDetailOnboardingStates, append(credentialOnboardingStates, shared.StateSuccess)...)...)
}

func (h ScheduleOnboardingStates) determineCredentialOnboardingStates(c flowpilot.HookExecutionContext, hasPasskey, hasPassword bool) []flowpilot.StateName {
	deps := h.GetDeps(c)
	cfg := deps.Cfg
	result := make([]flowpilot.StateName, 0)

	alwaysAcquirePasskey := cfg.Passkey.Enabled && cfg.Passkey.AcquireOnLogin == "always"
	alwaysAcquirePassword := cfg.Password.Enabled && cfg.Password.AcquireOnLogin == "always"
	conditionalAcquirePasskey := cfg.Passkey.Enabled && cfg.Passkey.AcquireOnLogin == "conditional"
	conditionalAcquirePassword := cfg.Password.Enabled && cfg.Password.AcquireOnLogin == "conditional"
	neverAcquirePasskey := !cfg.Passkey.Enabled || cfg.Passkey.AcquireOnLogin == "never"
	neverAcquirePassword := !cfg.Password.Enabled || cfg.Password.AcquireOnLogin == "never"

	if alwaysAcquirePasskey && alwaysAcquirePassword {
		if !hasPasskey && !hasPassword {
			if !cfg.Password.Optional && cfg.Passkey.Optional {
				result = append(result, shared.StatePasswordCreation, shared.StateOnboardingCreatePasskey)
			} else {
				result = append(result, shared.StateOnboardingCreatePasskey, shared.StatePasswordCreation)
			}
		} else if hasPasskey && !hasPassword {
			result = append(result, shared.StatePasswordCreation)
		} else if !hasPasskey && hasPassword {
			result = append(result, shared.StateOnboardingCreatePasskey)
		}
	} else if alwaysAcquirePasskey && conditionalAcquirePassword {
		if !hasPasskey && !hasPassword {
			result = append(result, shared.StateOnboardingCreatePasskey) // skip should lead to password onboarding
		} else if !hasPasskey && hasPassword {
			result = append(result, shared.StateOnboardingCreatePasskey)
		}
	} else if conditionalAcquirePasskey && alwaysAcquirePassword {
		if !hasPasskey && !hasPassword {
			result = append(result, shared.StatePasswordCreation) // skip should lead to passkey onboarding
		} else if hasPasskey && !hasPassword {
			result = append(result, shared.StatePasswordCreation)
		}
	} else if conditionalAcquirePasskey && conditionalAcquirePassword {
		if !hasPasskey && !hasPassword {
			result = append(result, shared.StateCredentialOnboardingChooser) // credential_onboarding_chooser can be skipped
		}
	} else if conditionalAcquirePasskey && neverAcquirePassword {
		if !hasPasskey && !hasPassword {
			result = append(result, shared.StateOnboardingCreatePasskey)
		}
	} else if neverAcquirePasskey && conditionalAcquirePassword {
		if !hasPasskey && !hasPassword {
			result = append(result, shared.StatePasswordCreation)
		}
	} else if neverAcquirePasskey && alwaysAcquirePassword {
		if !hasPassword {
			result = append(result, shared.StatePasswordCreation)
		}
	} else if alwaysAcquirePasskey && neverAcquirePassword {
		if !hasPasskey {
			result = append(result, shared.StateOnboardingCreatePasskey)
		}
	}

	return result
}

func (h ScheduleOnboardingStates) determineUserDetailOnboardingStates(c flowpilot.HookExecutionContext, userHasUsername, userHasEmail bool) []flowpilot.StateName {
	deps := h.GetDeps(c)
	cfg := deps.Cfg
	result := make([]flowpilot.StateName, 0)
	acquireUsername := !userHasUsername && cfg.Username.Enabled && cfg.Username.AcquireOnLogin
	acquireEmail := !userHasEmail && cfg.Email.Enabled && cfg.Email.AcquireOnLogin

	if acquireUsername && acquireEmail {
		result = append(result, shared.StateOnboardingUsername, shared.StateOnboardingEmail)
	} else if acquireUsername {
		result = append(result, shared.StateOnboardingUsername)
	} else if acquireEmail {
		result = append(result, shared.StateOnboardingEmail)
	}

	return result
}
