import {
  FetchStateFunction,
  MappedActions,
  MappedPayloads,
  StateName,
} from "./types/state-handling";
import { StateResponse } from "./types/state-response";
import { Error } from "./types/error";
import { Action } from "./types/action";

// State class represents a state in the flow
// eslint-disable-next-line require-jsdoc
class State<TStateName extends StateName> implements StateResponse<TStateName> {
  readonly name: StateName;
  readonly payload?: MappedPayloads[TStateName];
  readonly actions: MappedActions[TStateName];
  readonly error: Error;
  readonly status: number;

  private readonly fetchState: FetchStateFunction;

  // eslint-disable-next-line require-jsdoc
  constructor(
    stateResponse: StateResponse<TStateName>,
    fetchStateFunction: FetchStateFunction
  ) {
    this.name = stateResponse.name;
    this.payload = stateResponse.payload;
    this.actions = stateResponse.actions;
    this.error = stateResponse.error;
    this.status = stateResponse.status;

    this.fetchState = fetchStateFunction;
  }

  // Execute an action associated with this state
  async executeAction(action: Action<any>): Promise<State<any>> {
    const dataToSend: Record<string, any> = {};

    // eslint-disable-next-line guard-for-in
    for (const inputName in action.inputs) {
      dataToSend[inputName] = action.inputs[inputName]?.value;
    }

    // Use the fetch function to perform the action
    return this.fetchState(action.href, {
      input_data: JSON.stringify(dataToSend),
    });
  }
}

export { State };
