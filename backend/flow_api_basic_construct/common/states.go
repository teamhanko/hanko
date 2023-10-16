package common

import "github.com/teamhanko/hanko/backend/flowpilot"

const (
	StateSuccess flowpilot.StateName = "success"
	StateError   flowpilot.StateName = "error"

	StateLoginPreflight               flowpilot.StateName = "login_preflight"
	StateLoginInit                    flowpilot.StateName = "login_init"
	StateLoginMethodChooser           flowpilot.StateName = "login_method_chooser"
	StatePasswordLogin                flowpilot.StateName = "password_login"
	StateLoginPasscodeConfirmation    flowpilot.StateName = "login_passcode_confirmation"
	StateRecoveryPasscodeConfirmation flowpilot.StateName = "recovery_passcode_confirmation"
	StatePasskeyLogin                 flowpilot.StateName = "passkey_login"
	StateUse2FATOTP                   flowpilot.StateName = "use_2fa_totp"
	StateUse2FASecurityKey            flowpilot.StateName = "use_2fa_security_key"
	StateUseRecoveryCode              flowpilot.StateName = "use_recovery_code"
	StateRecoveryPasswordCreation     flowpilot.StateName = "recovery_password_creation"

	StateRegistrationPreflight flowpilot.StateName = "registration_preflight"
	StateRegistrationInit      flowpilot.StateName = "registration_init"
	StateEmailVerification     flowpilot.StateName = "registration_email_verification"
	StatePasswordCreation      flowpilot.StateName = "password_creation"

	StateOnboardingCreatePasskey            flowpilot.StateName = "onboarding_create_passkey"
	StateOnboardingVerifyPasskeyAttestation flowpilot.StateName = "onboarding_verify_passkey_attestation"

	StateCreate2FASecurityKey          flowpilot.StateName = "create_2fa_security_key"
	StateVerify2FASecurityKeyAssertion flowpilot.StateName = "verify_2fa_security_key_assertion"
	StateCreate2FATOTP                 flowpilot.StateName = "create_2fa_totp"
	StateGenerateRecoveryCodes         flowpilot.StateName = "generate_recovery_codes"
	StateShowRecoveryCodes             flowpilot.StateName = "show_recovery_codes"
)
