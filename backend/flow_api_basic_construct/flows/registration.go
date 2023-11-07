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
	"github.com/teamhanko/hanko/backend/session"
	"time"
)

func NewRegistrationFlow(cfg config.Config, persister persistence.Persister, passcodeService services.Passcode, sessionManager session.Manager, httpContext echo.Context) (flowpilot.Flow, error) {
	passkeyOnboardingSubFlow, err := NewPasskeyOnboardingSubFlow(cfg, persister)
	if err != nil {
		return nil, err
	}

	return flowpilot.NewFlow("/registration").
		State(common.StateRegistrationPreflight, actions.NewSendCapabilities(cfg)).
		State(common.StateRegistrationInit, actions.NewSubmitRegistrationIdentifier(cfg, persister, passcodeService, httpContext), actions.NewLoginWithOauth()).
		State(common.StateRegistrationPasscodeConfirmation, actions.NewSubmitPasscode(cfg, persister)).
		State(common.StatePasswordCreation, actions.NewSubmitNewPassword(cfg)).
		BeforeState(common.StateSuccess, hooks.NewBeforeSuccess(persister, sessionManager, httpContext)).
		State(common.StateSuccess).
		State(common.StateError).
		SubFlows(passkeyOnboardingSubFlow).
		InitialState(common.StateRegistrationPreflight).
		ErrorState(common.StateError).
		TTL(10 * time.Minute).
		Debug(true).
		MustBuild(), nil
}
