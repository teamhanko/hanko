package flows

import (
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/actions"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

func NewLoginFlow(cfg config.Config) flowpilot.Flow {
	return flowpilot.NewFlow("login").
		State(common.StatePreflight, actions.NewSendCapabilities(cfg)).
		State(common.StateLoginInit).
		State(common.StateLoginMethodChooser).
		State(common.StatePasskeyLogin).
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
		FixedStates(common.StatePreflight, common.StateError, common.StateSuccess).
		TTL(10 * time.Minute).
		MustBuild()
}
