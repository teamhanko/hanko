package registration

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/capabilities"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

const (
	StateRegistrationInit flowpilot.StateName = "registration_init"
	StatePasswordCreation flowpilot.StateName = "password_creation"
)

const (
	ActionRegisterPassword        flowpilot.ActionName = "register_password"
	ActionRegisterLoginIdentifier flowpilot.ActionName = "register_login_identifier"
)

var Flow = flowpilot.NewFlow("/registration").
	State(StateRegistrationInit, RegisterLoginIdentifier{}).
	State(StatePasswordCreation, RegisterPassword{}).
	BeforeState(shared.StateSuccess, CreateUser{}, shared.IssueSession{}).
	State(shared.StateSuccess).
	State(shared.StateError).
	SubFlows(capabilities.SubFlow, passkey_onboarding.SubFlow, passcode.SubFlow).
	InitialState(capabilities.StatePreflight, StateRegistrationInit).
	ErrorState(shared.StateError).
	TTL(10 * time.Minute).
	Debug(true).
	MustBuild()
