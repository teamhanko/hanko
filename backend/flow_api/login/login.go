package login

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/capabilities"
	capabilitiesStates "github.com/teamhanko/hanko/backend/flow_api/capabilities/states"
	"github.com/teamhanko/hanko/backend/flow_api/login/actions"
	"github.com/teamhanko/hanko/backend/flow_api/login/states"
	"github.com/teamhanko/hanko/backend/flow_api/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	sharedActions "github.com/teamhanko/hanko/backend/flow_api/shared/actions"
	"github.com/teamhanko/hanko/backend/flow_api/shared/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"time"
)

func NewLoginFlow(cfg config.Config, persister persistence.Persister, passcodeService services.Passcode, httpContext echo.Context) (flowpilot.Flow, error) {
	webauthn, err := cfg.Webauthn.GetConfig()
	if err != nil {
		return nil, err
	}

	onboardingSubFlow, err := passkey_onboarding.NewPasskeyOnboardingSubFlow(cfg, persister)
	if err != nil {
		return nil, err
	}

	capabilitiesSubFlow := capabilities.NewCapabilitiesSubFlow(cfg)

	passkeySubFlow, err := passcode.NewPasscodeSubFlow(cfg, persister, passcodeService, httpContext)
	if err != nil {
		return nil, err
	}

	return flowpilot.NewFlow("/login").
		State(states.StateLoginInit, actions.NewSubmitLoginIdentifier(cfg, persister, httpContext), sharedActions.NewLoginWithOauth(), actions.NewGetWARequestOptions(cfg, persister, webauthn)).
		State(states.StateLoginMethodChooser,
			actions.NewGetWARequestOptions(cfg, persister, webauthn),
			actions.NewLoginWithPassword(),
			actions.NewContinueToPasscodeConfirmation(cfg),
			sharedActions.NewBack(),
		).
		State(states.StateLoginPasskey, actions.NewSendWAAssertionResponse(cfg, persister, webauthn, httpContext)).
		State(states.StateLoginPassword,
			actions.NewSubmitPassword(cfg, persister),
			actions.NewContinueToPasscodeConfirmationRecovery(cfg),
			actions.NewContinueToLoginMethodChooser(),
			sharedActions.NewBack(),
		).
		State(states.StateLoginPasswordRecovery, sharedActions.NewSubmitNewPassword(cfg)).
		State(shared.StateSuccess).
		State(shared.StateError).
		SubFlows(capabilitiesSubFlow, onboardingSubFlow, passkeySubFlow).
		InitialState(capabilitiesStates.StatePreflight, states.StateLoginInit).
		ErrorState(shared.StateError).
		TTL(10 * time.Minute).
		MustBuild(), nil
}
