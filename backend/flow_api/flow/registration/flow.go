package registration

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/preflight"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

const (
	StateUserIdentificationPrompt flowpilot.StateName = "registration_user_identification_prompt"
	StateNewPasswordPrompt        flowpilot.StateName = "registration_password_prompt"
	StateSuccess                  flowpilot.StateName = "registration_success"
	StateError                    flowpilot.StateName = "registration_error"
)

const (
	ActionSetNewPassword    flowpilot.ActionName = "set_new_password"
	ActionSetUserIdentifier flowpilot.ActionName = "set_user_identifier"
)

var Flow = flowpilot.NewFlow("/registration").
	State(StateUserIdentificationPrompt, SetUserIdentifier{}).
	State(StateNewPasswordPrompt, SetNewPassword{}).
	State(StateSuccess).
	BeforeState(StateSuccess, CreateUser{}, shared.IssueSession{}).
	InitialState(preflight.StatePreflight, StateUserIdentificationPrompt).
	ErrorState(StateError).
	SubFlows(preflight.SubFlow, passkey_onboarding.SubFlow, passcode.SubFlow).
	TTL(10 * time.Minute).
	Debug(true).
	MustBuild()
