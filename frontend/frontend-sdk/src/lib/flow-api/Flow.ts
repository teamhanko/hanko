import { Client } from "../client/Client";
import { State, isState } from "./State";
import { Action } from "./types/action";
import { FetchNextState, FlowPath, Handlers } from "./types/state-handling";

// eslint-disable-next-line require-jsdoc
class Flow extends Client {
  public fetchNextState: FetchNextState = async (href: string, body?: any) => {
    const response = await this.client.post(href, body);
    return new State(response.json(), this.fetchNextState);
  };

  public async init(
    initPath: FlowPath,
    handlers: Handlers & { onError?: (e: unknown) => any }
  ): Promise<void> {
    const runLoop = async (path: string): Promise<void> => {
      const handlerResult = await this.run(path, handlers);

      // Handlers may return an action to be executed.
      // When the action is executed, we'll do all of this again (recursive),
      // so it looks somewhat like this:
      //
      // fetch next state -> handler -> action -> fetch next state -> handler -> action -> ...
      if (isAction(handlerResult)) {
        return runLoop(handlerResult.href);
      }
    };

    return runLoop(initPath);
  }

  /**
   * Runs a handler based on the current state and returns the result.
   *
   * @example
   * const handlerResult = await run("/login", {
   *   // all login handlers are in here, one of which will be called
   *   // based on what the /login endpoint returns
   * });
   */
  run = async (
    path: string,
    handlers: Handlers & { onError?: (e: unknown) => any }
  ) => {
    try {
      const state = await this.fetchNextState(path);

      if (!isState(state)) {
        throw new InvalidStateError(state);
      }

      const handler = handlers[state.name];
      if (!handler) {
        throw new HandlerNotFoundError(state);
      }

      return handler(state);
    } catch (e) {
      if (typeof handlers.onError === "function") {
        return handlers.onError(e);
      }

      throw e;
    }
  };
}

export class HandlerNotFoundError extends Error {
  constructor(public state: State<any>) {
    super(
      `No handler found for state: ${
        typeof state.name === "string"
          ? `"${state.name}"`
          : `(${typeof state.name})`
      }`
    );
  }
}

export class InvalidStateError extends Error {
  constructor(public state: State<any>) {
    super(
      `Invalid state: ${
        typeof state.name === "string"
          ? `"${state.name}"`
          : `(${typeof state.name})`
      }`
    );
  }
}

export function isAction(x: any): x is Action<unknown> {
  return typeof x === "object" && x !== null && "href" in x && "inputs" in x;
}

export { Flow };
