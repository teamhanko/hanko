package registration

import (
	"github.com/teamhanko/hanko/backend/flow_api/capabilities"
	"github.com/teamhanko/hanko/backend/flow_api/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flow_api/shared/hooks"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

const (
	StateRegistrationInit flowpilot.StateName = "registration_init"
	StatePasswordCreation flowpilot.StateName = "password_creation"
)

var Flow = flowpilot.NewFlow("/registration").
	State(StateRegistrationInit, SubmitRegistrationIdentifier{}).
	State(StatePasswordCreation, SubmitNewPassword{}).
	BeforeState(shared.StateSuccess, hooks.BeforeSuccess{}).
	State(shared.StateSuccess).
	State(shared.StateError).
	SubFlows(capabilities.SubFlow, passkey_onboarding.SubFlow, passcode.SubFlow).
	InitialState(capabilities.StatePreflight, StateRegistrationInit).
	ErrorState(shared.StateError).
	TTL(10 * time.Minute).
	Debug(true).
	MustBuild()
