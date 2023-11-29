package login

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/preflight"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

const (
	StateUserIdentificationPrompt flowpilot.StateName = "login_user_identification_prompt"
	StateMethodSelection          flowpilot.StateName = "login_method_selection"
	StatePasswordPrompt           flowpilot.StateName = "login_password_prompt"
	StateNewPasswordPrompt        flowpilot.StateName = "login_new_password_prompt"
	StatePasskeyAuthentication    flowpilot.StateName = "login_passkey_authentication"
	StateSuccess                  flowpilot.StateName = "login_success"
	StateError                    flowpilot.StateName = "login_error"
)

const (
	ActionContinueWithUserIdentifier             flowpilot.ActionName = "continue_with_user_identifier"
	ActionContinueToMethodSelection              flowpilot.ActionName = "continue_to_method_selection"
	ActionContinueToPasscodeConfirmationLogin    flowpilot.ActionName = "continue_to_passcode_confirmation_login"
	ActionContinueToPasscodeConfirmationRecovery flowpilot.ActionName = "continue_to_passcode_confirmation_recovery"
	ActionContinueToPasswordPrompt               flowpilot.ActionName = "continue_to_password_prompt"
	ActionWebauthnGenerateRequestOptions         flowpilot.ActionName = "webauthn_generate_request_options"
	ActionWebauthnVerifyAssertionResponse        flowpilot.ActionName = "webauthn_verify_request_response"
	ActionVerifyPassword                         flowpilot.ActionName = "verify_password"
	ActionSetNewPassword                         flowpilot.ActionName = "set_new_password"
)

var Flow = flowpilot.NewFlow("/login").
	State(StateUserIdentificationPrompt, ContinueWithUserIdentifier{}, WebauthnGenerateRequestOptions{}).
	State(StateMethodSelection,
		WebauthnGenerateRequestOptions{},
		ContinueToPasswordPrompt{},
		ContinueToPasscodeConfirmationLogin{},
		shared.Back{},
	).
	State(StatePasskeyAuthentication, WebauthnVerifyAssertionResponse{}).
	State(StatePasswordPrompt,
		VerifyPassword{},
		ContinueToPasscodeConfirmationForRecovery{},
		ContinueToMethodSelection{},
		shared.Back{},
	).
	State(StateNewPasswordPrompt, SetNewPassword{}).
	BeforeState(StateSuccess, shared.IssueSession{}).
	InitialState(preflight.StatePreflight, StateUserIdentificationPrompt).
	ErrorState(StateError).
	SubFlows(preflight.SubFlow, passkey_onboarding.SubFlow, passcode.SubFlow).
	TTL(10 * time.Minute).
	Debug(true).
	MustBuild()
