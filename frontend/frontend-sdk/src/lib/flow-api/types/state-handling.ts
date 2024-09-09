import { State } from "../State";

import {
  CredentialOnboardingChooserActions,
  LoginInitActions,
  LoginMethodChooserActions,
  LoginOTPActions,
  LoginPasskeyActions,
  LoginPasswordActions,
  LoginPasswordRecoveryActions,
  LoginSecurityKeyActions,
  MFAAOTPSecretCreationActions,
  MFAMethodChooserActions,
  MFASecurityKeyCreationActions,
  OnboardingCreatePasskeyActions,
  OnboardingEmailActions,
  OnboardingUsernameActions,
  OnboardingVerifyPasskeyAttestationActions,
  PasscodeConfirmationActions,
  PasswordCreationActions,
  PreflightActions,
  ProfileInitActions,
  RegistrationInitActions,
  ThirdPartyActions,
} from "./action";

import {
  LoginInitPayload,
  LoginPasskeyPayload,
  MFAOTPSecretCreationPayload,
  OnboardingVerifyPasskeyAttestationPayload,
  PasscodeConfirmationPayload,
  ProfilePayload,
  SuccessPayload,
  ThirdPartyPayload,
} from "./payload";

export type StateName =
  | "account_deleted"
  | "credential_onboarding_chooser"
  | "error"
  | "login_init"
  | "login_method_chooser"
  | "login_otp"
  | "login_passkey"
  | "login_password"
  | "login_password_recovery"
  | "login_security_key"
  | "mfa_method_chooser"
  | "mfa_otp_secret_creation"
  | "mfa_security_key_creation"
  | "onboarding_create_passkey"
  | "onboarding_email"
  | "onboarding_username"
  | "onboarding_verify_passkey_attestation"
  | "passcode_confirmation"
  | "password_creation"
  | "preflight"
  | "profile_init"
  | "registration_init"
  | "success"
  | "thirdparty"
  | "webauthn_credential_verification";

export interface Actions {
  readonly account_deleted: null;
  readonly credential_onboarding_chooser: CredentialOnboardingChooserActions;
  readonly error: null;
  readonly login_init: LoginInitActions;
  readonly login_method_chooser: LoginMethodChooserActions;
  readonly login_otp: LoginOTPActions;
  readonly login_passkey: LoginPasskeyActions;
  readonly login_password: LoginPasswordActions;
  readonly login_password_recovery: LoginPasswordRecoveryActions;
  readonly login_security_key: LoginSecurityKeyActions;
  readonly mfa_method_chooser: MFAMethodChooserActions;
  readonly mfa_otp_secret_creation: MFAAOTPSecretCreationActions;
  readonly mfa_security_key_creation: MFASecurityKeyCreationActions;
  readonly onboarding_create_passkey: OnboardingCreatePasskeyActions;
  readonly onboarding_email: OnboardingEmailActions;
  readonly onboarding_username: OnboardingUsernameActions;
  readonly onboarding_verify_passkey_attestation: OnboardingVerifyPasskeyAttestationActions;
  readonly passcode_confirmation: PasscodeConfirmationActions;
  readonly password_creation: PasswordCreationActions;
  readonly preflight: PreflightActions;
  readonly profile_init: ProfileInitActions;
  readonly registration_init: RegistrationInitActions;
  readonly success: null;
  readonly thirdparty: ThirdPartyActions;
  readonly webauthn_credential_verification: OnboardingVerifyPasskeyAttestationActions;
}

export interface Payloads {
  readonly account_deleted: null;
  readonly credential_onboarding_chooser: null;
  readonly error: null;
  readonly login_init: LoginInitPayload;
  readonly login_method_chooser: null;
  readonly login_otp: null;
  readonly login_passkey: LoginPasskeyPayload;
  readonly login_password: null;
  readonly login_password_recovery: null;
  readonly login_security_key: null;
  readonly mfa_method_chooser: null;
  readonly mfa_otp_secret_creation: MFAOTPSecretCreationPayload;
  readonly mfa_security_key_creation: null;
  readonly onboarding_create_passkey: null;
  readonly onboarding_email: null;
  readonly onboarding_username: null;
  readonly onboarding_verify_passkey_attestation: OnboardingVerifyPasskeyAttestationPayload;
  readonly passcode_confirmation: PasscodeConfirmationPayload;
  readonly password_creation: null;
  readonly preflight: null;
  readonly profile_init: ProfilePayload;
  readonly registration_init: null;
  readonly success: SuccessPayload;
  readonly thirdparty: ThirdPartyPayload;
  readonly webauthn_credential_verification: OnboardingVerifyPasskeyAttestationPayload;
}

export type FlowPath = "/login" | "/registration" | "/profile";

export type FetchNextState = (
  // eslint-disable-next-line no-unused-vars
  href: string,
  // eslint-disable-next-line no-unused-vars
  body?: any,
) => Promise<State<any>>;

export type HandlerFunction<TStateName extends StateName> = (
  // eslint-disable-next-line no-unused-vars
  state: State<TStateName>,
) => any;

export type Handlers = {
  [TStateName in StateName]: HandlerFunction<TStateName>;
};
