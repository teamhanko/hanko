import { State } from "../State";

import {
  Action,
  LoginInitActions,
  LoginMethodChooserActions,
  LoginPasskeyActions,
  LoginPasswordActions,
  LoginPasswordRecoveryActions,
  OnboardingCreatePasskeyActions,
  OnboardingVerifyPasskeyAttestationActions,
  PasscodeConfirmationActions,
  PasswordCreationActions,
  PreflightActions,
  RegistrationInitActions,
} from "./action";
import { Input } from "./input";

import {
  LoginPasskeyPayload,
  OnboardingVerifyPasskeyAttestationPayload,
} from "./payload";

export type StateName =
  | "preflight"
  | "login_init"
  | "login_method_chooser"
  | "login_password"
  | "login_password_recovery"
  | "passcode_confirmation"
  | "login_passkey"
  | "onboarding_create_passkey"
  | "onboarding_verify_passkey_attestation"
  | "registration_init"
  | "password_creation"
  | "success";

export interface Actions {
  readonly preflight: PreflightActions;
  readonly login_init: LoginInitActions;
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
}

export interface Payloads {
  readonly preflight: null;
  readonly login_init: null;
  readonly login_method_chooser: null;
  readonly login_password: null;
  readonly login_password_recovery: null;
  readonly passcode_confirmation: null;
  readonly login_passkey: LoginPasskeyPayload;
  readonly onboarding_create_passkey: null;
  readonly onboarding_verify_passkey_attestation: OnboardingVerifyPasskeyAttestationPayload;
  readonly registration_init: null;
  readonly password_creation: null;
  readonly success: null;
}

export type FlowPath = "/login" | "/registration" | "/profile";

export type FetchNextState = (
  // eslint-disable-next-line no-unused-vars
  href: string,
  // eslint-disable-next-line no-unused-vars
  body?: any
) => Promise<State<any>>;

export type HandlerFunction<TStateName extends StateName> = (
  // eslint-disable-next-line no-unused-vars
  state: State<TStateName>
) => any;

export type Handlers = {
  [TStateName in StateName]: HandlerFunction<TStateName>;
};
