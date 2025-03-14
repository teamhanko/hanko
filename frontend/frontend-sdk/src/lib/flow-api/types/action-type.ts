import {
  ContinueWithLoginIdentifierInputs,
  EmailCreateInputs,
  EmailDeleteInputs,
  EmailSetPrimaryInputs,
  EmailVerifyInputs,
  ExchangeTokenInputs,
  PasskeyCredentialDeleteInputs,
  PasskeyCredentialRenameInputs,
  PasswordRecoveryInputs,
  PasswordInputs,
  RegisterClientCapabilitiesInputs,
  RegisterLoginIdentifierInputs,
  RegisterPasswordInputs,
  ThirdpartyOauthInputs,
  UsernameSetInputs,
  VerifyPasscodeInputs,
  WebauthnVerifyAssertionResponseInputs,
  WebauthnVerifyAttestationResponseInputs,
  SessionDeleteInputs,
  OTPCodeInputs,
  SecurityKeyDeleteInputs,
  RememberMeInputs,
} from "./input";

export interface ActionType<TInputs> {
  action: string;
  href: string;
  inputs: TInputs;
  description: string;
}

export interface PreflightActions {
  readonly register_client_capabilities: ActionType<RegisterClientCapabilitiesInputs>;
}

export interface LoginInitActions {
  readonly continue_with_login_identifier?: ActionType<ContinueWithLoginIdentifierInputs>;
  readonly webauthn_generate_request_options?: ActionType<null>;
  readonly webauthn_verify_assertion_response?: ActionType<WebauthnVerifyAssertionResponseInputs>;
  readonly thirdparty_oauth?: ActionType<ThirdpartyOauthInputs>;
  readonly remember_me?: ActionType<RememberMeInputs>;
}

export interface ProfileInitActions {
  readonly account_delete?: ActionType<null>;
  readonly continue_to_otp_secret_creation?: ActionType<null>;
  readonly email_create?: ActionType<EmailCreateInputs>;
  readonly email_delete?: ActionType<EmailDeleteInputs>;
  readonly email_verify?: ActionType<EmailVerifyInputs>;
  readonly email_set_primary?: ActionType<EmailSetPrimaryInputs>;
  readonly otp_secret_delete?: ActionType<null>;
  readonly password_create?: ActionType<PasswordInputs>;
  readonly password_update?: ActionType<PasswordInputs>;
  readonly password_delete?: ActionType<null>;
  readonly security_key_create?: ActionType<null>;
  readonly security_key_delete?: ActionType<SecurityKeyDeleteInputs>;
  readonly username_create?: ActionType<UsernameSetInputs>;
  readonly username_delete?: ActionType<null>;
  readonly username_update?: ActionType<UsernameSetInputs>;
  readonly webauthn_credential_create?: ActionType<null>;
  readonly webauthn_credential_rename?: ActionType<PasskeyCredentialRenameInputs>;
  readonly webauthn_credential_delete?: ActionType<PasskeyCredentialDeleteInputs>;
  readonly webauthn_verify_attestation_response?: ActionType<WebauthnVerifyAttestationResponseInputs>;
  readonly session_delete?: ActionType<SessionDeleteInputs>;
}

export interface LoginMethodChooserActions {
  readonly continue_to_password_login?: ActionType<null>;
  readonly continue_to_passcode_confirmation?: ActionType<null>;
  readonly back: ActionType<null>;
}

export interface LoginOTPActions {
  readonly otp_code_validate: ActionType<OTPCodeInputs>;
  readonly continue_to_login_security_key?: ActionType<null>;
}

export interface LoginPasswordActions {
  readonly password_login: ActionType<PasswordInputs>;
  readonly continue_to_passcode_confirmation_recovery?: ActionType<null>;
  readonly continue_to_login_method_chooser: ActionType<null>;
  readonly back: ActionType<null>;
}

export interface LoginPasswordRecoveryActions {
  readonly password_recovery: ActionType<PasswordRecoveryInputs>;
}

export interface LoginPasskeyActions {
  readonly webauthn_verify_assertion_response: ActionType<WebauthnVerifyAssertionResponseInputs>;
  readonly back: ActionType<null>;
}

export interface LoginSecurityKeyActions {
  readonly webauthn_generate_request_options: ActionType<null>;
  readonly continue_to_login_otp?: ActionType<null>;
}

export interface MFAMethodChooserActions {
  readonly continue_to_otp_secret_creation?: ActionType<null>;
  readonly continue_to_security_key_creation?: ActionType<null>;
  readonly skip?: ActionType<null>;
  readonly back?: ActionType<null>;
}

export interface MFAAOTPSecretCreationActions {
  readonly otp_code_verify: ActionType<OTPCodeInputs>;
  readonly back: ActionType<null>;
}

export interface MFASecurityKeyCreationActions {
  readonly webauthn_generate_creation_options: ActionType<null>;
  readonly back: ActionType<null>;
}

export interface OnboardingCreatePasskeyActions {
  readonly webauthn_generate_creation_options: ActionType<null>;
  readonly skip?: ActionType<null>;
  readonly back?: ActionType<null>;
}

export interface OnboardingVerifyPasskeyAttestationActions {
  readonly webauthn_verify_attestation_response: ActionType<WebauthnVerifyAttestationResponseInputs>;
  readonly back: ActionType<null>;
}

export interface RegistrationInitActions {
  readonly register_login_identifier: ActionType<RegisterLoginIdentifierInputs>;
  readonly thirdparty_oauth?: ActionType<ThirdpartyOauthInputs>;
  readonly remember_me?: ActionType<RememberMeInputs>;
}

export interface PasswordCreationActions {
  readonly register_password: ActionType<RegisterPasswordInputs>;
  readonly back?: ActionType<null>;
  readonly skip?: ActionType<null>;
}

export interface PasscodeConfirmationActions {
  readonly verify_passcode: ActionType<VerifyPasscodeInputs>;
  readonly resend_passcode: ActionType<null>;
  readonly back: ActionType<null>;
}

export interface OnboardingEmailActions {
  readonly email_address_set: ActionType<EmailCreateInputs>;
  readonly skip: ActionType<null>;
}

export interface OnboardingUsernameActions {
  readonly username_create: ActionType<UsernameSetInputs>;
  readonly skip: ActionType<null>;
}

export interface CredentialOnboardingChooserActions {
  readonly continue_to_passkey_registration: ActionType<null>;
  readonly continue_to_password_registration: ActionType<null>;
  readonly skip: ActionType<null>;
  readonly back: ActionType<null>;
}

export interface DeviceTrustActions {
  readonly trust_device: ActionType<null>;
  readonly skip: ActionType<null>;
  readonly back?: ActionType<null>;
}

export interface ThirdPartyActions {
  readonly exchange_token: ActionType<ExchangeTokenInputs>;
}
