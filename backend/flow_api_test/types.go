package flow_api_test

import "github.com/teamhanko/hanko/backend/flowpilot"

const (
	StateSignInOrSignUp             flowpilot.StateName = "init"
	StateError                      flowpilot.StateName = "error"
	StateSuccess                    flowpilot.StateName = "success"
	StateConfirmAccountCreation     flowpilot.StateName = "confirmation"
	StateLoginWithPassword          flowpilot.StateName = "login_with_password"
	StateLoginWithPasscode          flowpilot.StateName = "login_with_passcode"
	StateLoginWithPasskey           flowpilot.StateName = "login_with_passkey"
	StateCreatePasskey              flowpilot.StateName = "create_passkey"
	StateUpdateExistingPassword     flowpilot.StateName = "update_existing_password"
	StateRecoverPasswordViaPasscode flowpilot.StateName = "recover_password_via_passcode"
	StatePasswordCreation           flowpilot.StateName = "password_creation"
	StateConfirmPasskeyCreation     flowpilot.StateName = "confirm_passkey_creation"
	StateVerifyEmailViaPasscode     flowpilot.StateName = "verify_email_via_passcode"
)

const (
	MethodSubmitEmail            flowpilot.MethodName = "submit_email"
	MethodGetWAChallenge         flowpilot.MethodName = "get_webauthn_challenge"
	MethodVerifyWAPublicKey      flowpilot.MethodName = "verify_webauthn_public_key"
	MethodGetWAAssertion         flowpilot.MethodName = "get_webauthn_assertion"
	MethodVerifyWAAssertion      flowpilot.MethodName = "verify_webauthn_assertion_response"
	MethodSubmitExistingPassword flowpilot.MethodName = "submit_existing_password"
	MethodSubmitNewPassword      flowpilot.MethodName = "submit_new_password"
	MethodRequestRecovery        flowpilot.MethodName = "request_recovery"
	MethodSubmitPasscodeCode     flowpilot.MethodName = "submit_passcode_code"
	MethodCreateUser             flowpilot.MethodName = "create_user"
	MethodSkipPasskeyCreation    flowpilot.MethodName = "skip_passkey_creation"
	MethodBack                   flowpilot.MethodName = "back"
)
