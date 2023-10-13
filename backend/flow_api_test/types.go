package flow_api_test

import "github.com/teamhanko/hanko/backend/flowpilot"

const (
	StateSecondSubFlowInit  flowpilot.StateName = "StateSecondSubFlowInit"
	StateThirdSubFlowInit   flowpilot.StateName = "StateThirdSubFlowInit"
	StateSecondSubFlowFinal flowpilot.StateName = "StateSecondSubFlowFinal"
	StateFirstSubFlowInit   flowpilot.StateName = "StateFirstSubFlowInit"

	StateSignInOrSignUp             flowpilot.StateName = "init"
	StateError                      flowpilot.StateName = "error"
	StateSuccess                    flowpilot.StateName = "success"
	StateConfirmAccountCreation     flowpilot.StateName = "confirmation"
	StateLoginWithPassword          flowpilot.StateName = "login_with_password"
	StateLoginWithPasscode          flowpilot.StateName = "login_with_passcode"
	StateLoginWithPasscode2FA       flowpilot.StateName = "login_with_passcode_2fa"
	StateLoginWithPasskey           flowpilot.StateName = "login_with_passkey"
	StateCreatePasskey              flowpilot.StateName = "create_passkey"
	StateUpdateExistingPassword     flowpilot.StateName = "update_existing_password"
	StateRecoverPasswordViaPasscode flowpilot.StateName = "recover_password_via_passcode"
	StatePasswordCreation           flowpilot.StateName = "password_creation"
	StateConfirmPasskeyCreation     flowpilot.StateName = "confirm_passkey_creation"
	StateVerifyEmailViaPasscode     flowpilot.StateName = "verify_email_via_passcode"
)

const (
	ActionEndSubFlow         flowpilot.ActionName = "EndSubFlow"
	ActionContinueToFinal    flowpilot.ActionName = "ContinueToFinal"
	ActionStartFirstSubFlow  flowpilot.ActionName = "StartFirstSubFlow"
	ActionStartSecondSubFlow flowpilot.ActionName = "StartSecondSubFlow"
	ActionStartThirdSubFlow  flowpilot.ActionName = "StartThirdSubFlow"

	ActionSubmitEmail            flowpilot.ActionName = "submit_email"
	ActionGetWAChallenge         flowpilot.ActionName = "get_webauthn_challenge"
	ActionVerifyWAPublicKey      flowpilot.ActionName = "verify_webauthn_public_key"
	ActionGetWAAssertion         flowpilot.ActionName = "get_webauthn_assertion"
	ActionVerifyWAAssertion      flowpilot.ActionName = "verify_webauthn_assertion_response"
	ActionSubmitExistingPassword flowpilot.ActionName = "submit_existing_password"
	ActionSubmitNewPassword      flowpilot.ActionName = "submit_new_password"
	ActionRequestRecovery        flowpilot.ActionName = "request_recovery"
	ActionSubmitPasscodeCode     flowpilot.ActionName = "submit_passcode_code"
	ActionCreateUser             flowpilot.ActionName = "create_user"
	ActionSkipPasskeyCreation    flowpilot.ActionName = "skip_passkey_creation"
	ActionBack                   flowpilot.ActionName = "back"
)
