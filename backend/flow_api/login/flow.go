package login

import (
	"github.com/teamhanko/hanko/backend/flow_api/capabilities"
	"github.com/teamhanko/hanko/backend/flow_api/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/registration"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	sharedActions "github.com/teamhanko/hanko/backend/flow_api/shared/actions"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

const (
	StateLoginInit             flowpilot.StateName = "login_init"
	StateLoginMethodChooser    flowpilot.StateName = "login_method_chooser"
	StateLoginPassword         flowpilot.StateName = "login_password"
	StateLoginPasskey          flowpilot.StateName = "login_passkey"
	StateLoginPasswordRecovery flowpilot.StateName = "login_password_recovery"
)

var Flow = flowpilot.NewFlow("/login").
	State(StateLoginInit, SubmitLoginIdentifier{}, GetWARequestOptions{}).
	State(StateLoginMethodChooser,
		GetWARequestOptions{},
		LoginWithPassword{},
		ContinueToPasscodeConfirmation{},
		sharedActions.Back{},
	).
	State(StateLoginPasskey, SendWAAssertionResponse{}).
	State(StateLoginPassword,
		SubmitPassword{},
		ContinueToPasscodeConfirmationRecovery{},
		ContinueToLoginMethodChooser{},
		sharedActions.Back{},
	).
	State(StateLoginPasswordRecovery, registration.SubmitNewPassword{}).
	State(shared.StateSuccess).
	State(shared.StateError).
	SubFlows(capabilities.SubFlow, passkey_onboarding.SubFlow, passcode.SubFlow).
	InitialState(capabilities.StatePreflight, StateLoginInit).
	ErrorState(shared.StateError).
	TTL(10 * time.Minute).
	MustBuild()
