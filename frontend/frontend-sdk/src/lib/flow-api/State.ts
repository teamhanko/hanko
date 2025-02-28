import { Actions, Payloads, StateName } from "./types/state-handling";
import { Error } from "./types/error";
import { Action } from "./types/action";
import { Input } from "./types/input";

// Derived AllStates from StateName
export type AllStates = { [K in StateName]: State<K> }[StateName];

// FetchFunction returns State<StateName> since next state is dynamic
export type FetchFunction = (
  // eslint-disable-next-line no-unused-vars
  href: string,
  // eslint-disable-next-line no-unused-vars
  body?: any,
) => Promise<AllStates>;

// Helper types
type ExtractInputValues<TInputs> = {
  [K in keyof TInputs]: TInputs[K] extends Input<infer TValue> ? TValue : never;
};

interface ActionHandler<TInputs> {
  enabled: boolean;
  inputs: TInputs;
  // eslint-disable-next-line no-unused-vars
  run(userInputs: ExtractInputValues<TInputs>): Promise<AllStates>;
}

type ActionMap<TState extends StateName> = {
  [K in keyof Actions[TState]]: ActionHandler<
    Actions[TState][K] extends Action<infer TInputs> ? TInputs : never
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

// eslint-disable-next-line require-jsdoc
export class State<TState extends StateName = StateName> {
  public readonly name: TState;
  public readonly error?: Error;
  public readonly payload?: Payloads[TState];
  public readonly actions: ActionMap<TState>;
  private readonly csrfToken: string;
  private readonly fetchFunc: FetchFunction;

  // eslint-disable-next-line require-jsdoc
  constructor(response: FlowResponse<TState>, fetchState: FetchFunction) {
    this.name = response.name; // No cast needed
    this.error = response.error;
    this.payload = response.payload;
    this.csrfToken = response.csrf_token;
    this.actions = this.buildActions(response.actions);
    this.fetchFunc = fetchState;
  }

  // eslint-disable-next-line require-jsdoc
  private buildActions(actions: Actions[TState]): ActionMap<TState> {
    const actionMap: ActionMap<TState> = {} as any;

    Object.keys(actions).forEach((actionName) => {
      const key = actionName as keyof Actions[TState];
      const action = actions[key] as Action<any>;
      actionMap[key] = {
        enabled: action.enabled,
        inputs: action.inputs,
        run: async (inputValues: ExtractInputValues<typeof action.inputs>) => {
          if (!action.enabled) {
            throw new Error(
              `Action '${String(key)}' is not enabled in state '${this.name}'`,
            );
          }
          return this.executeAction(action.href, inputValues);
        },
      };
    });

    return actionMap;
  }

  // eslint-disable-next-line require-jsdoc
  private async executeAction(
    href: string,
    inputValues: Record<string, any>,
  ): Promise<AllStates> {
    const requestBody = {
      input_data: inputValues,
      csrf_token: this.csrfToken,
    };
    return this.fetchFunc(href, requestBody);
  }
}
