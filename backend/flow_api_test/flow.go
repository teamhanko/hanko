package flow_api_test

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

var ThirdSubFlow = flowpilot.NewSubFlow().
	State(StateThirdSubFlowInit, EndSubFlow{}, Back{}).
	MustBuild()

var SecondSubFlow = flowpilot.NewSubFlow().
	State(StateSecondSubFlowInit, ContinueToFinal{}, Back{}).
	State(StateSecondSubFlowFinal, EndSubFlow{}, Back{}).
	AfterState(StateSecondSubFlowFinal, AfterStateAction{}).
	MustBuild()

var FirstSubFlow = flowpilot.NewSubFlow().
	State(StateFirstSubFlowInit, StartSecondSubFlow{}, Back{}).
	SubFlows(SecondSubFlow).
	MustBuild()

var Flow = flowpilot.NewFlow("/flow_api_login").
	State(StateSignInOrSignUp, SubmitEmail{}, GetWAChallenge{}, StartFirstSubFlow{}, Back{}).
	State(StateLoginWithPasskey, VerifyWAPublicKey{}, Back{}).
	State(StateLoginWithPasscode, SubmitPasscodeCode{}, Back{}).
	State(StateLoginWithPasscode2FA, SubmitPasscodeCode{}).
	State(StateRecoverPasswordViaPasscode, SubmitPasscodeCode{}, Back{}).
	State(StateVerifyEmailViaPasscode, SubmitPasscodeCode{}).
	State(StateLoginWithPassword, SubmitExistingPassword{}, RequestRecovery{}, Back{}).
	State(StateUpdateExistingPassword, SubmitNewPassword{}).
	State(StateConfirmAccountCreation, CreateUser{}, Back{}).
	State(StatePasswordCreation, SubmitNewPassword{}).
	State(StateConfirmPasskeyCreation, GetWAAssertion{}, SkipPasskeyCreation{}).
	State(StateCreatePasskey, VerifyWAAssertion{}).
	BeforeState(StateSuccess, BeforeStateAction{}).
	InitialState(StateSignInOrSignUp).
	ErrorState(StateError).
	SubFlows(FirstSubFlow, ThirdSubFlow).
	TTL(time.Minute * 10).
	Debug(true).
	MustBuild()
