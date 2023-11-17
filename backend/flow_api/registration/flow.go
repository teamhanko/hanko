package registration

import (
	"github.com/teamhanko/hanko/backend/flow_api/capabilities"
	capabilitiesStates "github.com/teamhanko/hanko/backend/flow_api/capabilities/states"
	"github.com/teamhanko/hanko/backend/flow_api/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/registration/actions"
	"github.com/teamhanko/hanko/backend/flow_api/registration/states"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	sharedActions "github.com/teamhanko/hanko/backend/flow_api/shared/actions"
	"github.com/teamhanko/hanko/backend/flow_api/shared/hooks"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

var Flow = flowpilot.NewFlow("/registration").
	State(states.StateRegistrationInit, actions.SubmitRegistrationIdentifier{}, sharedActions.LoginWithOauth{}).
	State(shared.StatePasswordCreation, sharedActions.SubmitNewPassword{}).
	BeforeState(shared.StateSuccess, hooks.BeforeSuccess{}).
	State(shared.StateSuccess).
	State(shared.StateError).
	SubFlows(capabilities.SubFlow, passkey_onboarding.SubFlow, passcode.SubFlow).
	InitialState(capabilitiesStates.StatePreflight, states.StateRegistrationInit).
	ErrorState(shared.StateError).
	TTL(10 * time.Minute).
	Debug(true).
	MustBuild()
