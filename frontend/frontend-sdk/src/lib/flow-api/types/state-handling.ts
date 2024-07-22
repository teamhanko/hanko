import { State } from "../State";

import {
  CredentialOnboardingChooserActions,
  LoginInitActions,
  LoginMethodChooserActions,
  LoginPasskeyActions,
  LoginPasswordActions,
  LoginPasswordRecoveryActions,
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
  OnboardingVerifyPasskeyAttestationPayload,
  PasscodeConfirmationPayload,
  ProfilePayload,
  ThirdPartyPayload,
  SuccessPayload,
} from "./payload";

export type StateName =
  | "error"
  | "login_init"
  | "login_method_chooser"
  | "login_passkey"
  | "login_password"
  | "login_password_recovery"
  | "onboarding_create_passkey"
  | "onboarding_verify_passkey_attestation"
  | "passcode_confirmation"
  | "password_creation"
  | "preflight"
  | "profile_init"
  | "registration_init"
  | "success"
  | "webauthn_credential_verification"
  | "onboarding_email"
  | "onboarding_username"
  | "credential_onboarding_chooser"
  | "thirdparty";

export interface Actions {
  readonly preflight: PreflightActions;
  readonly login_init: LoginInitActions;
  readonly profile_init: ProfileInitActions;
  readonly webauthn_credential_verification: OnboardingVerifyPasskeyAttestationActions;
  readonly login_method_chooser: LoginMethodChooserActions;
  readonly login_password: LoginPasswordActions;
  readonly login_password_recovery: LoginPasswordRecoveryActions;
  readonly passcode_confirmation: PasscodeConfirmationActions;
  readonly login_passkey: LoginPasskeyActions;
  readonly onboarding_create_passkey: OnboardingCreatePasskeyActions;
  readonly onboarding_verify_passkey_attestation: OnboardingVerifyPasskeyAttestationActions;
  readonly registration_init: RegistrationInitActions;
  readonly password_creation: PasswordCreationActions;
  readonly success: null;
  readonly error: null;
  readonly onboarding_email: OnboardingEmailActions;
  readonly onboarding_username: OnboardingUsernameActions;
  readonly credential_onboarding_chooser: CredentialOnboardingChooserActions;
  readonly thirdparty: ThirdPartyActions;
}

export interface Payloads {
  readonly preflight: null;
  readonly login_init: LoginInitPayload;
  readonly profile_init: ProfilePayload;
  readonly webauthn_credential_verification: OnboardingVerifyPasskeyAttestationPayload;
  readonly login_method_chooser: null;
  readonly login_password: null;
  readonly login_password_recovery: null;
  readonly passcode_confirmation: PasscodeConfirmationPayload;
  readonly login_passkey: LoginPasskeyPayload;
  readonly onboarding_create_passkey: null;
  readonly onboarding_verify_passkey_attestation: OnboardingVerifyPasskeyAttestationPayload;
  readonly registration_init: null;
  readonly password_creation: null;
  readonly success: SuccessPayload;
  readonly error: null;
  readonly onboarding_email: null;
  readonly onboarding_username: null;
  readonly credential_onboarding_chooser: null;
  readonly thirdparty: ThirdPartyPayload;
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
