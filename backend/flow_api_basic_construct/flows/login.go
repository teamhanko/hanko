package flows

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/actions"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/hooks"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"time"
)

func NewLoginFlow(cfg config.Config, persister persistence.Persister, passcodeService services.Passcode, httpContext echo.Context) (flowpilot.Flow, error) {
	webauthn, err := getWebauthn(cfg)
	if err != nil {
		return nil, err
	}

	onboardingSubFlow, err := NewPasskeyOnboardingSubFlow(cfg, persister)
	if err != nil {
		return nil, err
	}

	capabilitiesSubFlow := NewCapabilitiesSubFlow(cfg)

	return flowpilot.NewFlow("/login").
		State(common.StateLoginInit, actions.NewSubmitLoginIdentifier(cfg, persister, httpContext), actions.NewLoginWithOauth(), actions.NewGetWARequestOptions(cfg, persister, webauthn)).
		State(common.StateLoginMethodChooser,
			actions.NewGetWARequestOptions(cfg, persister, webauthn),
			actions.NewLoginWithPassword(),
			actions.NewContinueToPasscodeConfirmation(cfg),
			actions.NewBack(),
		).
		State(common.StateLoginPasskey, actions.NewSendWAAssertionResponse(cfg, persister, webauthn, httpContext)).
		State(common.StateLoginPassword,
			actions.NewSubmitPassword(cfg, persister),
			actions.NewContinueToPasscodeConfirmationRecovery(cfg),
			actions.NewContinueToLoginMethodChooser(),
			actions.NewBack(),
		).
		State(common.StateLoginPasscodeConfirmationRecovery, actions.NewSubmitPasscode(cfg, persister)).
		State(common.StateLoginPasscodeConfirmation, actions.NewSubmitPasscode(cfg, persister)).
		//State(common.StateUse2FASecurityKey).
		//State(common.StateUse2FATOTP).
		//State(common.StateUseRecoveryCode).
		State(common.StateLoginPasswordRecovery, actions.NewSubmitNewPassword(cfg)).
		State(common.StateSuccess).
		State(common.StateError).
		BeforeState(common.StateLoginPasscodeConfirmation, hooks.NewSendPasscode(passcodeService, httpContext)).
		BeforeState(common.StateLoginPasscodeConfirmationRecovery, hooks.NewSendPasscode(passcodeService, httpContext)).
		SubFlows(capabilitiesSubFlow, onboardingSubFlow).
		InitialState(common.StatePreflight, common.StateLoginInit).
		ErrorState(common.StateError).
		TTL(10 * time.Minute).
		MustBuild(), nil
}
