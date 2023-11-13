package common

import "github.com/teamhanko/hanko/backend/flowpilot"

const (
	StateSuccess flowpilot.StateName = "success"
	StateError   flowpilot.StateName = "error"

	StatePreflight flowpilot.StateName = "preflight"

	StateLoginInit                         flowpilot.StateName = "login_init"
	StateLoginMethodChooser                flowpilot.StateName = "login_method_chooser"
	StateLoginPassword                     flowpilot.StateName = "login_password"
	StateLoginPasscodeConfirmation         flowpilot.StateName = "login_passcode_confirmation"
	StateLoginPasscodeConfirmationRecovery flowpilot.StateName = "login_passcode_confirmation_recovery"
	StateLoginPasskey                      flowpilot.StateName = "login_passkey"
	StateUse2FATOTP                        flowpilot.StateName = "use_2fa_totp"
	StateUse2FASecurityKey                 flowpilot.StateName = "use_2fa_security_key"
	StateUseRecoveryCode                   flowpilot.StateName = "use_recovery_code"
	StateLoginPasswordRecovery             flowpilot.StateName = "login_password_recovery"

	StateRegistrationInit                 flowpilot.StateName = "registration_init"
	StateRegistrationPasscodeConfirmation flowpilot.StateName = "registration_passcode_confirmation"
	StatePasswordCreation                 flowpilot.StateName = "password_creation"

	StateOnboardingCreatePasskey            flowpilot.StateName = "onboarding_create_passkey"
	StateOnboardingVerifyPasskeyAttestation flowpilot.StateName = "onboarding_verify_passkey_attestation"

	StateCreate2FASecurityKey          flowpilot.StateName = "create_2fa_security_key"
	StateVerify2FASecurityKeyAssertion flowpilot.StateName = "verify_2fa_security_key_assertion"
	StateCreate2FATOTP                 flowpilot.StateName = "create_2fa_totp"
	StateGenerateRecoveryCodes         flowpilot.StateName = "generate_recovery_codes"
	StateShowRecoveryCodes             flowpilot.StateName = "show_recovery_codes"
)
