package flow_api_test

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

var Flow = flowpilot.NewFlow("/flow_api_login").
	State(StateSignInOrSignUp, SubmitEmail{}, GetWAChallenge{}).
	State(StateLoginWithPasskey, VerifyWAPublicKey{}, Back{}).
	State(StateLoginWithPasscode, SubmitPasscodeCode{}, Back{}).
	State(StateLoginWithPassword, SubmitExistingPassword{}, RequestRecovery{}, Back{}).
	State(StateRecoverPasswordViaPasscode, SubmitPasscodeCode{}, Back{}).
	State(StateUpdateExistingPassword, SubmitNewPassword{}).
	State(StateConfirmAccountCreation, CreateUser{}, Back{}).
	State(StatePasswordCreation, SubmitNewPassword{}).
	State(StateConfirmPasskeyCreation, GetWAAssertion{}, SkipPasskeyCreation{}).
	State(StateCreatePasskey, VerifyWAAssertion{}).
	State(StateVerifyEmailViaPasscode, SubmitPasscodeCode{}).
	State(StateError).
	State(StateSuccess).
	FixedStates(StateSignInOrSignUp, StateError, StateSuccess).
	TTL(time.Minute * 10).
	Build()
