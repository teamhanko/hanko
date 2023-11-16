package shared

import "github.com/teamhanko/hanko/backend/flowpilot"

const (
	StateSuccess flowpilot.StateName = "success"
	StateError   flowpilot.StateName = "error"

	StateUse2FATOTP        flowpilot.StateName = "use_2fa_totp"
	StateUse2FASecurityKey flowpilot.StateName = "use_2fa_security_key"
	StateUseRecoveryCode   flowpilot.StateName = "use_recovery_code"

	StatePasswordCreation flowpilot.StateName = "password_creation"

	StateCreate2FASecurityKey          flowpilot.StateName = "create_2fa_security_key"
	StateVerify2FASecurityKeyAssertion flowpilot.StateName = "verify_2fa_security_key_assertion"
	StateCreate2FATOTP                 flowpilot.StateName = "create_2fa_totp"
	StateGenerateRecoveryCodes         flowpilot.StateName = "generate_recovery_codes"
	StateShowRecoveryCodes             flowpilot.StateName = "show_recovery_codes"
)
