import { StateName, Actions, Payloads } from "./state";
import { ActionType } from "./actionType";
import { Input } from "./input";
import { Error } from "./error";
import { State } from "../State";
import { Action } from "../Action";

export type FlowPath = "/login" | "/registration" | "/profile";

export type AllStates = { [K in StateName]: State<K> }[StateName];

// eslint-disable-next-line no-unused-vars
export type FetchFunction = (href: string, body?: any) => Promise<AllStates>;

export type ExtractInputValues<TInputs> = {
  [K in keyof TInputs]: TInputs[K] extends Input<infer TValue> ? TValue : never;
};

export type AutoStep<TState extends StateName> = (
  // eslint-disable-next-line no-unused-vars
  state: State<TState>,
) => Promise<AllStates>;

export type DefaultHandler<TState extends StateName> = (
  // eslint-disable-next-line no-unused-vars
  state: State<TState>,
) => Promise<void>;

type PickStates<T extends StateName> = T;

export type AutoSteps = {
  [TStateName in PickStates<
    | "preflight"
    | "login_passkey"
    | "onboarding_verify_passkey_attestation"
    | "webauthn_credential_verification"
  >]: AutoStep<TStateName>;
};

export type DefaultHandlers = {
  [TStateName in PickStates<
    "login_init" | "success"
  >]: DefaultHandler<TStateName>;
};

export type ActionMap<TState extends StateName> = {
  [K in keyof Actions[TState]]: Action<
    Actions[TState][K] extends ActionType<infer TInputs> ? TInputs : never
  >;
};

export interface FlowResponse<TState extends StateName> {
  name: TState;
  status: number;
  payload?: Payloads[TState];
  actions?: Actions[TState];
  csrf_token: string;
  error?: Error;
}
