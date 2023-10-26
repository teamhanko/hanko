package flows

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/actions"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
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

	return flowpilot.NewFlow("/login").
		State(common.StateLoginPreflight, actions.NewSendCapabilities(cfg)).
		State(common.StateLoginInit, actions.NewSubmitLoginIdentifier(cfg, persister, httpContext), actions.NewLoginWithOauth(), actions.NewGetWARequestOptions(cfg, persister, webauthn)).
		State(common.StateLoginMethodChooser, actions.NewGetWARequestOptions(cfg, persister, webauthn), actions.NewLoginWithPassword(), actions.NewSendPasscodeCode(passcodeService, httpContext)).
		State(common.StatePasskeyLogin, actions.NewSendWAAssertionResponse(cfg, persister, webauthn, httpContext)).
		State(common.StatePasswordLogin).
		State(common.StateRecoveryPasscodeConfirmation).
		State(common.StateLoginPasscodeConfirmation).
		State(common.StateUse2FASecurityKey).
		State(common.StateUse2FATOTP).
		State(common.StateUseRecoveryCode).
		State(common.StateRecoveryPasswordCreation).
		State(common.StateSuccess).
		State(common.StateError).
		//SubFlows(NewPasskeyOnboardingSubFlow(), New2FACreationSubFlow()).
		InitialState(common.StateLoginPreflight).
		ErrorState(common.StateError).
		TTL(10 * time.Minute).
		MustBuild(), nil
}
