import { Client } from "../client/Client";
import { State, isState } from "./State";
import { Action } from "./types/action";
import { FetchNextState, FlowPath, Handlers } from "./types/state-handling";

type MaybePromise<T> = T | Promise<T>;

// eslint-disable-next-line require-jsdoc
class Flow extends Client {
  public fetchNextState: FetchNextState = async (href: string, body?: any) => {
    const response = await this.client.post(href, body);
    return new State(response.json(), this.fetchNextState);
  };

  private handlers: (Handlers & { onError?: (e: unknown) => any }) | undefined;

  public async init(
    initPath: FlowPath,
    handlers: Handlers & { onError?: (e: unknown) => any },
    getInitState: (flow: Flow) => MaybePromise<State<any> | null> = () =>
      this.fetchNextState(initPath)
  ): Promise<void> {
    this.handlers = handlers;

    const initState = await getInitState(this);

    await this.run(initState);
  }

  /**
   * Runs a handler for a given state.
   *
   * If the handler returns an action or a state, this method will run the next
   * appropriate handler for that state. (Recursively)
   *
   * If the handlers passed to `init` do not contain an `onError` handler,
   * this method will throw.
   *
   * @see InvalidStateError
   * @see HandlerNotFoundError
   *
   * @example
   * const handlerResult = await run("/login", {
   *   // all login handlers are in here, one of which will be called
   *   // based on what the /login endpoint returns
   * });
   */
  run = async (state: State<any>): Promise<unknown> => {
    try {
      if (!isState(state)) {
        throw new InvalidStateError(state);
      }

      const handler = this.handlers[state.name];
      if (!handler) {
        throw new HandlerNotFoundError(state);
      }

      let maybeNextState = await handler(state);

      // handler can return an action, which we'll run (and turn into state)...
      if (isAction(maybeNextState)) {
        maybeNextState = await (maybeNextState as any).run();
      }

      // ...or a state, to continue the "run loop"
      if (isState(maybeNextState)) {
        return this.run(maybeNextState);
      }
    } catch (e) {
      if (typeof this.handlers.onError === "function") {
        return this.handlers.onError(e);
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
