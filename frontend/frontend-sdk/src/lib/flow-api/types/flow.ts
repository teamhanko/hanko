import { StateName, Actions, Payloads } from "./state";
import { FlowError } from "./flowError";
import { State } from "../State";

type PickStates<TState extends StateName> = TState;

export type FlowName = "login" | "registration" | "profile" | "token_exchange";

export type AnyState = { [TState in StateName]: State<TState> }[StateName];

export type AutoStep<TState extends StateName> = (
  // eslint-disable-next-line no-unused-vars
  state: State<TState>,
) => Promise<AnyState>;

export type AutoSteps = {
  [TState in PickStates<
    | "preflight"
    | "login_passkey"
    | "onboarding_verify_passkey_attestation"
    | "webauthn_credential_verification"
    | "thirdparty"
    | "success"
    | "account_deleted"
  >]: AutoStep<TState>;
};

export type PasskeyAutofillActivationHandler<TState extends StateName> = (
  // eslint-disable-next-line no-unused-vars
  state: State<TState>,
) => Promise<void>;

export type PasskeyAutofillActivationHandlers = {
  [TState in PickStates<"login_init">]: PasskeyAutofillActivationHandler<TState>;
};

export interface FlowResponse<TState extends StateName> {
  name: TState;
  status: number;
  payload?: Payloads[TState];
  actions?: Actions[TState];
  csrf_token: string;
  error?: FlowError;
}
