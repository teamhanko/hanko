import { Client } from "../client/Client";
import { State, isState } from "./State";
import { Action } from "./types/action";
import { FetchNextState, FlowPath, Handlers } from "./types/state-handling";

type MaybePromise<T> = T | Promise<T>;

type ExtendedHandlers = Handlers & { onError?: (e: unknown) => any };
type GetInitState = (flow: Flow) => MaybePromise<State<any> | null>;

// eslint-disable-next-line require-jsdoc
class Flow extends Client {
  public async init(
    initPath: FlowPath,
    handlers: ExtendedHandlers,
    // getInitState: GetInitState = () => this.fetchNextState(initPath),
  ): Promise<void> {
    const fetchNextState: FetchNextState = async (href: string, body?: any) => {
      try {
        const response = await this.client.post(href, body);
        return new State(response.json(), fetchNextState);
      } catch (e) {
        handlers.onError?.(e);
      }
    };

    const initState = await fetchNextState(initPath);
    await this.run(initState, handlers);
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
  run = async (
    state: State<any>,
    handlers: ExtendedHandlers,
  ): Promise<unknown> => {
    try {
      if (!isState(state)) {
        throw new InvalidStateError(state);
      }

      const handler = handlers[state.name];
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
        return this.run(maybeNextState, handlers);
      }
    } catch (e) {
      if (typeof handlers.onError === "function") {
        return handlers.onError(e);
      }
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
      }`,
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
      }`,
    );
  }
}

export function isAction(x: any): x is Action<unknown> {
  return typeof x === "object" && x !== null && "href" in x && "inputs" in x;
}

export { Flow };
