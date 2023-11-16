package registration

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/capabilities"
	capabilitiesStates "github.com/teamhanko/hanko/backend/flow_api/capabilities/states"
	"github.com/teamhanko/hanko/backend/flow_api/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/registration/actions"
	"github.com/teamhanko/hanko/backend/flow_api/registration/states"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	sharedActions "github.com/teamhanko/hanko/backend/flow_api/shared/actions"
	"github.com/teamhanko/hanko/backend/flow_api/shared/hooks"
	"github.com/teamhanko/hanko/backend/flow_api/shared/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/session"
	"time"
)

func NewRegistrationFlow(cfg config.Config, persister persistence.Persister, passcodeService services.Passcode, sessionManager session.Manager, httpContext echo.Context) (flowpilot.Flow, error) {
	passkeyOnboardingSubFlow, err := passkey_onboarding.NewPasskeyOnboardingSubFlow(cfg, persister)
	if err != nil {
		return nil, err
	}

	capabilitiesSubFlow := capabilities.NewCapabilitiesSubFlow(cfg)

	passcodeSubFlow, err := passcode.NewPasscodeSubFlow(cfg, persister, passcodeService, httpContext)
	if err != nil {
		return nil, err
	}

	return flowpilot.NewFlow("/registration").
		State(states.StateRegistrationInit, actions.NewSubmitRegistrationIdentifier(cfg, persister, passcodeService, httpContext), sharedActions.NewLoginWithOauth()).
		State(shared.StatePasswordCreation, sharedActions.NewSubmitNewPassword(cfg)).
		BeforeState(shared.StateSuccess, hooks.NewBeforeSuccess(persister, sessionManager, httpContext)).
		State(shared.StateSuccess).
		State(shared.StateError).
		SubFlows(capabilitiesSubFlow, passkeyOnboardingSubFlow, passcodeSubFlow).
		InitialState(capabilitiesStates.StatePreflight, states.StateRegistrationInit).
		ErrorState(shared.StateError).
		TTL(10 * time.Minute).
		Debug(true).
		MustBuild(), nil
}
