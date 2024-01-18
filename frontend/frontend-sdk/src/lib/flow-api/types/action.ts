import {
  ContinueWithLoginIdentifierInputs,
  PasswordLoginInputs,
  PasswordRecoveryInputs,
  RegisterClientCapabilitiesInputs,
  RegisterLoginIdentifierInputs,
  RegisterPasswordInputs,
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
}

interface OnboardingCreatePasskeyActions {
  readonly webauthn_generate_creation_options: Action<null>;
  readonly skip?: Action<null>;
}

interface OnboardingVerifyPasskeyAttestationActions {
  readonly webauthn_verify_attestation_response: Action<WebauthnVerifyAttestationResponseInputs>;
}

interface RegistrationInitActions {
  readonly register_login_identifier: Action<RegisterLoginIdentifierInputs>;
}

interface PasswordCreationActions {
  readonly register_password: Action<RegisterPasswordInputs>;
}

interface PasscodeConfirmationActions {
  readonly verify_passcode: Action<VerifyPasscodeInputs>;
  readonly resend_passcode: Action<null>;
}

export type {
  Action,
  PreflightActions,
  LoginInitActions,
  LoginMethodChooserActions,
  LoginPasswordActions,
  LoginPasswordRecoveryActions,
  LoginPasskeyActions,
  OnboardingCreatePasskeyActions,
  OnboardingVerifyPasskeyAttestationActions,
  RegistrationInitActions,
  PasswordCreationActions,
  PasscodeConfirmationActions,
};
