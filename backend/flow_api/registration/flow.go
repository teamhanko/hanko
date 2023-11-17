package registration

import (
	"github.com/teamhanko/hanko/backend/flow_api/capabilities"
	"github.com/teamhanko/hanko/backend/flow_api/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	sharedActions "github.com/teamhanko/hanko/backend/flow_api/shared/actions"
	"github.com/teamhanko/hanko/backend/flow_api/shared/hooks"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

const (
	StateRegistrationInit flowpilot.StateName = "registration_init"
)

var Flow = flowpilot.NewFlow("/registration").
	State(StateRegistrationInit, SubmitRegistrationIdentifier{}).
	State(shared.StatePasswordCreation, sharedActions.SubmitNewPassword{}).
	BeforeState(shared.StateSuccess, hooks.BeforeSuccess{}).
	State(shared.StateSuccess).
	State(shared.StateError).
	SubFlows(capabilities.SubFlow, passkey_onboarding.SubFlow, passcode.SubFlow).
	InitialState(capabilities.StatePreflight, StateRegistrationInit).
	ErrorState(shared.StateError).
	TTL(10 * time.Minute).
	Debug(true).
	MustBuild()
