package login

import (
	"github.com/teamhanko/hanko/backend/flow_api/capabilities"
	capabilitiesStates "github.com/teamhanko/hanko/backend/flow_api/capabilities/states"
	"github.com/teamhanko/hanko/backend/flow_api/login/actions"
	"github.com/teamhanko/hanko/backend/flow_api/login/states"
	"github.com/teamhanko/hanko/backend/flow_api/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	sharedActions "github.com/teamhanko/hanko/backend/flow_api/shared/actions"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

var Flow = flowpilot.NewFlow("/login").
	State(states.StateLoginInit, actions.SubmitLoginIdentifier{}, sharedActions.LoginWithOauth{}, actions.GetWARequestOptions{}).
	State(states.StateLoginMethodChooser,
		actions.GetWARequestOptions{},
		actions.LoginWithPassword{},
		actions.ContinueToPasscodeConfirmation{},
		sharedActions.Back{},
	).
	State(states.StateLoginPasskey, actions.SendWAAssertionResponse{}).
	State(states.StateLoginPassword,
		actions.SubmitPassword{},
		actions.ContinueToPasscodeConfirmationRecovery{},
		actions.ContinueToLoginMethodChooser{},
		sharedActions.Back{},
	).
	State(states.StateLoginPasswordRecovery, sharedActions.SubmitNewPassword{}).
	State(shared.StateSuccess).
	State(shared.StateError).
	SubFlows(capabilities.SubFlow, passkey_onboarding.SubFlow, passcode.SubFlow).
	InitialState(capabilitiesStates.StatePreflight, states.StateLoginInit).
	ErrorState(shared.StateError).
	TTL(10 * time.Minute).
	MustBuild()
