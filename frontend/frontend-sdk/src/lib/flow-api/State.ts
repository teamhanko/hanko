import {
  Actions,
  FetchState,
  Payloads,
  StateName,
} from "./types/state-handling";

import { Error } from "./types/error";
import { Input } from "./types/input";
import { Action } from "./types/action";

interface StateResponse<TStateName extends StateName> {
  name: StateName;
  status: number;
  payload?: Payloads[TStateName];
  actions?: Actions[TStateName];
  csrf_token: string;
  error: Error;
}

type InputValues<TInput extends Record<string, Input<any>>> = {
  [K in keyof TInput]?: TInput[K]["value"];
};

// eslint-disable-next-line require-jsdoc
class WrappedAction<TInputs> {
  readonly action: Action<TInputs>;
  readonly fetchState: FetchState;
  // eslint-disable-next-line require-jsdoc
  constructor(action: Action<TInputs>, fetchState: FetchState) {
    this.action = action;
    this.fetchState = fetchState;
  }

  // eslint-disable-next-line require-jsdoc
  run(inputs: InputValues<TInputs>): Promise<State<any>> {
    return this.fetchState(this.action.href, {
      inputs: { ...this.action.inputs, inputs },
    });
  }
}

type WrappedActions<TStateName extends StateName> = {
  // eslint-disable-next-line no-unused-vars
  [name in keyof Actions[TStateName]]: WrappedAction<TStateName>;
};

// eslint-disable-next-line require-jsdoc
export class State<TStateName extends StateName>
  implements Omit<StateResponse<TStateName>, "actions">
{
  readonly name: StateName;
  readonly payload?: Payloads[TStateName];
  readonly error: Error;
  readonly status: number;
  readonly csrf_token: string;
  readonly actions: WrappedActions<TStateName>;

  // eslint-disable-next-line require-jsdoc
  constructor(
    {
      name,
      payload,
      error,
      status,
      actions,
      // eslint-disable-next-line camelcase
      csrf_token,
    }: StateResponse<TStateName>,
    fetchState: FetchState,
  ) {
    this.name = name;
    this.payload = payload;
    this.error = error;
    this.status = status;
    // eslint-disable-next-line camelcase
    this.csrf_token = csrf_token;

    for (const name in actions) {
      if (Object.prototype.hasOwnProperty.call(actions, name)) {
        const action = actions[name];
        this.actions[name] = new WrappedAction<TStateName>(action, fetchState);
      }
    }
  }
}
