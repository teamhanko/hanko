import { StateName, Actions, Payloads } from "./state";
import { ActionType as ActionType } from "./actionType";
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

type PickStates<T extends StateName> = T;

export type AutoSteps = {
  [TStateName in PickStates<"preflight">]: AutoStep<TStateName>;
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
