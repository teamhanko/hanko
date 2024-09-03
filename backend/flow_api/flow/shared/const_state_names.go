package shared

import "github.com/teamhanko/hanko/backend/flowpilot"

const (
	StateError                                 flowpilot.StateName = "error"
	StateLoginInit                             flowpilot.StateName = "login_init"
	StateLoginMethodChooser                    flowpilot.StateName = "login_method_chooser"
	StateLoginPasskey                          flowpilot.StateName = "login_passkey"
	StateLoginPassword                         flowpilot.StateName = "login_password"
	StateLoginPasswordRecovery                 flowpilot.StateName = "login_password_recovery"
	StateLoginSecurityKey                      flowpilot.StateName = "login_security_key"
	StateLoginOTP                              flowpilot.StateName = "login_otp"
	StateOnboardingCreatePasskey               flowpilot.StateName = "onboarding_create_passkey"
	StateCredentialOnboardingChooser           flowpilot.StateName = "credential_onboarding_chooser"
	StateOnboardingVerifyPasskeyAttestation    flowpilot.StateName = "onboarding_verify_passkey_attestation"
	StatePasscodeConfirmation                  flowpilot.StateName = "passcode_confirmation"
	StatePasswordCreation                      flowpilot.StateName = "password_creation"
	StatePreflight                             flowpilot.StateName = "preflight"
	StateProfileAccountDeleted                 flowpilot.StateName = "account_deleted"
	StateProfileInit                           flowpilot.StateName = "profile_init"
	StateProfileWebauthnCredentialVerification flowpilot.StateName = "webauthn_credential_verification"
	StateRegistrationInit                      flowpilot.StateName = "registration_init"
	StateSuccess                               flowpilot.StateName = "success"
	StateThirdParty                            flowpilot.StateName = "thirdparty"
	StateOnboardingEmail                       flowpilot.StateName = "onboarding_email"
	StateOnboardingUsername                    flowpilot.StateName = "onboarding_username"
)
