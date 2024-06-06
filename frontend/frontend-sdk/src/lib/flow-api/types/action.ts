import {
  ContinueWithLoginIdentifierInputs,
  EmailCreateInputs,
  EmailDeleteInputs,
  EmailSetPrimaryInputs,
  EmailVerifyInputs, ExchangeTokenInputs,
  PasskeyCredentialDelete,
  PasskeyCredentialRename,
  PasswordLoginInputs,
  PasswordRecoveryInputs,
  PasswordSetInputs,
  RegisterClientCapabilitiesInputs,
  RegisterLoginIdentifierInputs,
  RegisterPasswordInputs, ThirdpartyOauthInputs,
  UsernameSetInputs,
  VerifyPasscodeInputs,
  WebauthnVerifyAssertionResponseInputs,
  WebauthnVerifyAttestationResponseInputs,
} from "./input";

interface Action<TInputs> {
  name: string;
  href: string;
  inputs: TInputs;
  description: string;
}

interface PreflightActions {
  readonly register_client_capabilities: Action<RegisterClientCapabilitiesInputs>;
}

interface LoginInitActions {
  readonly continue_with_login_identifier: Action<ContinueWithLoginIdentifierInputs>;
  readonly webauthn_generate_request_options?: Action<null>;
  readonly webauthn_verify_assertion_response?: Action<WebauthnVerifyAssertionResponseInputs>;
  readonly thirdparty_oauth?: Action<ThirdpartyOauthInputs>;
}

interface ProfileInitActions {
  readonly account_delete?: Action<null>;
  readonly email_create?: Action<EmailCreateInputs>;
  readonly email_delete?: Action<EmailDeleteInputs>;
  readonly email_verify?: Action<EmailVerifyInputs>;
  readonly email_set_primary?: Action<EmailSetPrimaryInputs>;
  readonly password_set?: Action<PasswordSetInputs>;
  readonly password_delete?: Action<null>;
  readonly username_set?: Action<UsernameSetInputs>;
  readonly webauthn_credential_create?: Action<null>;
  readonly webauthn_credential_rename?: Action<PasskeyCredentialRename>;
  readonly webauthn_credential_delete?: Action<PasskeyCredentialDelete>;
  readonly webauthn_verify_attestation_response?: Action<WebauthnVerifyAttestationResponseInputs>;
}

interface LoginMethodChooserActions {
  readonly webauthn_generate_request_options?: Action<null>;
  readonly continue_to_password_login?: Action<null>;
  readonly continue_to_passcode_confirmation?: Action<null>;
  readonly back: Action<null>;
}

interface LoginPasswordActions {
  readonly password_login: Action<PasswordLoginInputs>;
  readonly continue_to_passcode_confirmation_recovery?: Action<null>;
  readonly continue_to_login_method_chooser: Action<null>;
  readonly back: Action<null>;
}

interface LoginPasswordRecoveryActions {
  readonly password_recovery: Action<PasswordRecoveryInputs>;
}

interface LoginPasskeyActions {
  readonly webauthn_verify_assertion_response: Action<WebauthnVerifyAssertionResponseInputs>;
  readonly back: Action<null>;
}

interface OnboardingCreatePasskeyActions {
  readonly webauthn_generate_creation_options: Action<null>;
  readonly skip?: Action<null>;
}

interface OnboardingVerifyPasskeyAttestationActions {
  readonly webauthn_verify_attestation_response: Action<WebauthnVerifyAttestationResponseInputs>;
  readonly back: Action<null>;
}

interface RegistrationInitActions {
  readonly register_login_identifier: Action<RegisterLoginIdentifierInputs>;
  readonly thirdparty_oauth?: Action<ThirdpartyOauthInputs>;
}

interface PasswordCreationActions {
  readonly register_password: Action<RegisterPasswordInputs>;
  readonly back?: Action<null>;
}

interface PasscodeConfirmationActions {
  readonly verify_passcode: Action<VerifyPasscodeInputs>;
  readonly resend_passcode: Action<null>;
  readonly back: Action<null>;
}

interface ThirdpartyOauthActions {
  readonly exchange_token: Action<ExchangeTokenInputs>;
}

export type {
  Action,
  PreflightActions,
  LoginInitActions,
  ProfileInitActions,
  LoginMethodChooserActions,
  LoginPasswordActions,
  LoginPasswordRecoveryActions,
  LoginPasskeyActions,
  OnboardingCreatePasskeyActions,
  OnboardingVerifyPasskeyAttestationActions,
  RegistrationInitActions,
  PasswordCreationActions,
  PasscodeConfirmationActions,
  ThirdpartyOauthActions
};
