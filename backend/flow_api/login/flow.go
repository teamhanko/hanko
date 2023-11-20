package login

import (
	"github.com/teamhanko/hanko/backend/flow_api/capabilities"
	"github.com/teamhanko/hanko/backend/flow_api/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/registration"
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
	ActionGetWARequestOptions                    flowpilot.ActionName = "get_wa_request_options"
	ActionLoginWithPassword                      flowpilot.ActionName = "login_with_password"
	ActionSendWAAssertionResponse                flowpilot.ActionName = "send_wa_request_response"
	ActionSubmitLoginIdentifier                  flowpilot.ActionName = "submit_login_identifier"
	ActionRecoverPassword                        flowpilot.ActionName = "recover_password"
	ActionSubmitPassword                         flowpilot.ActionName = "submit_password"
)

var Flow = flowpilot.NewFlow("/login").
	State(StateLoginInit, SubmitLoginIdentifier{}, GetWARequestOptions{}).
	State(StateLoginMethodChooser,
		GetWARequestOptions{},
		LoginWithPassword{},
		ContinueToPasscodeConfirmation{},
		shared.Back{},
	).
	State(StateLoginPasskey, SendWAAssertionResponse{}).
	State(StateLoginPassword,
		SubmitPassword{},
		ContinueToPasscodeConfirmationRecovery{},
		ContinueToLoginMethodChooser{},
		shared.Back{},
	).
	State(StateLoginPasswordRecovery, registration.SubmitNewPassword{}).
	State(shared.StateSuccess).
	State(shared.StateError).
	SubFlows(capabilities.SubFlow, passkey_onboarding.SubFlow, passcode.SubFlow).
	InitialState(capabilities.StatePreflight, StateLoginInit).
	ErrorState(shared.StateError).
	TTL(10 * time.Minute).
	MustBuild()
