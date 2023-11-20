package login

import (
	"github.com/teamhanko/hanko/backend/flow_api/capabilities"
	"github.com/teamhanko/hanko/backend/flow_api/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
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

const (
	ActionContinueToLoginMethodChooser           flowpilot.ActionName = "continue_to_login_method_chooser"
	ActionContinueToPasscodeConfirmation         flowpilot.ActionName = "continue_to_passcode_confirmation"
	ActionContinueToPasscodeConfirmationRecovery flowpilot.ActionName = "continue_to_passcode_confirmation_recovery"
	ActionContinueToPasswordLogin                flowpilot.ActionName = "continue_to_password_login"
	ActionWebauthnGenerateRequestOptions         flowpilot.ActionName = "webauthn_generate_request_options"
	ActionWebauthnVerifyAssertionResponse        flowpilot.ActionName = "webauthn_verify_request_response"
	ActionContinueWithLoginIdentifier            flowpilot.ActionName = "continue_with_login_identifier"
	ActionPasswordRecovery                       flowpilot.ActionName = "password_recovery"
	ActionPasswordLogin                          flowpilot.ActionName = "password_login"
)

var Flow = flowpilot.NewFlow("/login").
	State(StateLoginInit, ContinueWithLoginIdentifier{}, WebauthnGenerateRequestOptions{}).
	State(StateLoginMethodChooser,
		WebauthnGenerateRequestOptions{},
		ContinueToPasswordLogin{},
		ContinueToPasscodeConfirmation{},
		shared.Back{},
	).
	State(StateLoginPasskey, WebauthnVerifyAssertionResponse{}).
	State(StateLoginPassword,
		PasswordLogin{},
		ContinueToPasscodeConfirmationRecovery{},
		ContinueToLoginMethodChooser{},
		shared.Back{},
	).
	State(StateLoginPasswordRecovery, PasswordRecovery{}).
	State(shared.StateSuccess).
	State(shared.StateError).
	SubFlows(capabilities.SubFlow, passkey_onboarding.SubFlow, passcode.SubFlow).
	InitialState(capabilities.StatePreflight, StateLoginInit).
	ErrorState(shared.StateError).
	TTL(10 * time.Minute).
	MustBuild()
