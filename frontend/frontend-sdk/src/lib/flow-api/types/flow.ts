import { StateName, Actions, Payloads } from "./state";
import { FlowError } from "./flowError";
import { State } from "../State";

type PickStates<TState extends StateName> = TState;

/**
 * The name of a top-level flow exposed by the flow API. Each flow is a self-contained state
 * machine (e.g. logging in, registering, or managing a profile).
 * @category SDK
 * @subcategory Flow Types
 */
export type FlowName = "login" | "registration" | "profile" | "token_exchange";

/**
 * A {@link State} narrowed to any one of the possible {@link StateName}s a flow can be in.
 * @category SDK
 * @subcategory Flow Types
 */
export type AnyState = { [TState in StateName]: State<TState> }[StateName];

/**
 * A handler that automatically runs a follow-up action for a given state, without requiring
 * explicit user interaction. See {@link State#autoStep}.
 * @template TState - The state name the handler applies to.
 * @category SDK
 * @subcategory Flow Types
 */
export type AutoStep<TState extends StateName> = (
  // eslint-disable-next-line no-unused-vars
  state: State<TState>,
) => Promise<AnyState>;

/**
 * Maps each state that supports an automatic step (see {@link AutoStep}) to its handler.
 * @category SDK
 * @subcategory Flow Types
 */
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

/**
 * A handler that activates WebAuthn passkey autofill (conditional mediation) for a given state.
 * See {@link State#passkeyAutofillActivation}.
 * @template TState - The state name the handler applies to.
 * @category SDK
 * @subcategory Flow Types
 */
export type PasskeyAutofillActivationHandler<TState extends StateName> = (
  // eslint-disable-next-line no-unused-vars
  state: State<TState>,
) => Promise<void>;

/**
 * Maps each state that supports passkey autofill activation (see {@link PasskeyAutofillActivationHandler})
 * to its handler.
 * @category SDK
 * @subcategory Flow Types
 */
export type PasskeyAutofillActivationHandlers = {
  [TState in PickStates<"login_init">]: PasskeyAutofillActivationHandler<TState>;
};

/**
 * The raw response returned by the flow API backend for a given state, before it is wrapped in a
 * {@link State} instance.
 * @template TState - The specific state name type.
 * @interface
 * @category SDK
 * @subcategory Flow Types
 * @property {TState} name - The name of the state.
 * @property {number} status - The HTTP status code of the response.
 * @property {*} [payload] - The state's payload data, if any (see {@link Payloads}).
 * @property {*} [actions] - The actions available from this state (see {@link Actions}).
 * @property {string} csrf_token - The CSRF token to include when running an action from this state.
 * @property {FlowError} [error] - A general error for this state, if the previous action failed.
 */
export interface FlowResponse<TState extends StateName> {
  name: TState;
  status: number;
  payload?: Payloads[TState];
  actions?: Actions[TState];
  csrf_token: string;
  error?: FlowError;
}
