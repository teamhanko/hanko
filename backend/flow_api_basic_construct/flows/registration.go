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

func NewRegistrationFlow(cfg config.Config, persister persistence.Persister, passcodeService *services.Passcode, httpContext echo.Context) flowpilot.Flow {
	// TODO:
	return flowpilot.NewFlow("/registration").
		State(common.StatePreflight, actions.NewSendCapabilities(cfg)).
		State(common.StateRegistrationInit, actions.NewSubmitRegistrationIdentifier(cfg, persister, passcodeService, httpContext), actions.NewLoginWithOauth()).
		State(common.StateEmailVerification, actions.NewSubmitPasscode()).
		State(common.StatePasswordCreation).
		State(common.StateSuccess).
		State(common.StateError).
		//SubFlows(NewPasskeyOnboardingSubFlow(), New2FACreationSubFlow()).
		FixedStates(common.StatePreflight, common.StateError, common.StateSuccess).
		TTL(10 * time.Minute).
		Debug(true).
		MustBuild()
}
