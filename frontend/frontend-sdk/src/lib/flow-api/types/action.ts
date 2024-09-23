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
  OTPCodeInputs,
  SecurityKeyDeleteInputs,
} from "./input";

export interface Action<TInputs> {
  name: string;
  href: string;
  inputs: TInputs;
  description: string;
}

export interface PreflightActions {
  readonly register_client_capabilities: Action<RegisterClientCapabilitiesInputs>;
}

export interface LoginInitActions {
  readonly continue_with_login_identifier?: Action<ContinueWithLoginIdentifierInputs>;
  readonly webauthn_generate_request_options?: Action<null>;
  readonly webauthn_verify_assertion_response?: Action<WebauthnVerifyAssertionResponseInputs>;
  readonly thirdparty_oauth?: Action<ThirdpartyOauthInputs>;
}

export interface ProfileInitActions {
  readonly account_delete?: Action<null>;
  readonly continue_to_otp_secret_creation?: Action<null>;
  readonly email_create?: Action<EmailCreateInputs>;
  readonly email_delete?: Action<EmailDeleteInputs>;
  readonly email_verify?: Action<EmailVerifyInputs>;
  readonly email_set_primary?: Action<EmailSetPrimaryInputs>;
  readonly otp_secret_delete?: Action<null>;
  readonly password_create?: Action<PasswordInputs>;
  readonly password_update?: Action<PasswordInputs>;
  readonly password_delete?: Action<null>;
  readonly security_key_create?: Action<null>;
  readonly security_key_delete?: Action<SecurityKeyDeleteInputs>;
  readonly username_create?: Action<UsernameSetInputs>;
  readonly username_delete?: Action<null>;
  readonly username_update?: Action<UsernameSetInputs>;
  readonly webauthn_credential_create?: Action<null>;
  readonly webauthn_credential_rename?: Action<PasskeyCredentialRenameInputs>;
  readonly webauthn_credential_delete?: Action<PasskeyCredentialDeleteInputs>;
  readonly webauthn_verify_attestation_response?: Action<WebauthnVerifyAttestationResponseInputs>;
}

export interface LoginMethodChooserActions {
  readonly continue_to_password_login?: Action<null>;
  readonly continue_to_passcode_confirmation?: Action<null>;
  readonly back: Action<null>;
}

export interface LoginOTPActions {
  readonly otp_code_validate: Action<OTPCodeInputs>;
  readonly continue_to_login_security_key?: Action<null>;
}

export interface LoginPasswordActions {
  readonly password_login: Action<PasswordInputs>;
  readonly continue_to_passcode_confirmation_recovery?: Action<null>;
  readonly continue_to_login_method_chooser: Action<null>;
  readonly back: Action<null>;
}

export interface LoginPasswordRecoveryActions {
  readonly password_recovery: Action<PasswordRecoveryInputs>;
}

export interface LoginPasskeyActions {
  readonly webauthn_verify_assertion_response: Action<WebauthnVerifyAssertionResponseInputs>;
  readonly back: Action<null>;
}

export interface LoginSecurityKeyActions {
  readonly webauthn_generate_request_options: Action<null>;
  readonly continue_to_login_otp?: Action<null>;
}

export interface MFAMethodChooserActions {
  readonly continue_to_otp_secret_creation?: Action<null>;
  readonly continue_to_security_key_creation?: Action<null>;
  readonly skip?: Action<null>;
  readonly back?: Action<null>;
}

export interface MFAAOTPSecretCreationActions {
  readonly otp_code_verify: Action<OTPCodeInputs>;
  readonly back: Action<null>;
}

export interface MFASecurityKeyCreationActions {
  readonly webauthn_generate_creation_options: Action<null>;
  readonly back: Action<null>;
}

export interface OnboardingCreatePasskeyActions {
  readonly webauthn_generate_creation_options: Action<null>;
  readonly skip?: Action<null>;
  readonly back?: Action<null>;
}

export interface OnboardingVerifyPasskeyAttestationActions {
  readonly webauthn_verify_attestation_response: Action<WebauthnVerifyAttestationResponseInputs>;
  readonly back: Action<null>;
}

export interface RegistrationInitActions {
  readonly register_login_identifier: Action<RegisterLoginIdentifierInputs>;
  readonly thirdparty_oauth?: Action<ThirdpartyOauthInputs>;
}

export interface PasswordCreationActions {
  readonly register_password: Action<RegisterPasswordInputs>;
  readonly back?: Action<null>;
  readonly skip?: Action<null>;
}

export interface PasscodeConfirmationActions {
  readonly verify_passcode: Action<VerifyPasscodeInputs>;
  readonly resend_passcode: Action<null>;
  readonly back: Action<null>;
}

export interface OnboardingEmailActions {
  readonly email_address_set: Action<EmailCreateInputs>;
  readonly skip: Action<null>;
}

export interface OnboardingUsernameActions {
  readonly username_create: Action<UsernameSetInputs>;
  readonly skip: Action<null>;
}

export interface CredentialOnboardingChooserActions {
  readonly continue_to_passkey_registration: Action<null>;
  readonly continue_to_password_registration: Action<null>;
  readonly skip: Action<null>;
  readonly back: Action<null>;
}

export interface ThirdPartyActions {
  readonly exchange_token: Action<ExchangeTokenInputs>;
}
