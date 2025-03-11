import { Actions, Payloads, StateName } from "./types/state";
import { Error } from "./types/error";
import { ActionType } from "./types/actionType";
import { ActionMap, AnyState, FlowPath, FlowResponse } from "./types/flow";
import { Action } from "./Action";
import { autoSteps, defaultHandlers } from "./auto-steps";
import { Hanko } from "../../Hanko";

type AutoSteppedStates = keyof typeof autoSteps;
type DefaultHandledStates = keyof typeof defaultHandlers;
type SerializedState = FlowResponse<any> & { path: FlowPath };

// eslint-disable-next-line require-jsdoc
export class State<TState extends StateName = StateName> {
  public readonly name: TState;
  public readonly path: FlowPath;
  public error?: Error;
  public readonly payload?: Payloads[TState];
  public readonly actions: ActionMap<TState>;
  private readonly csrfToken: string;
  private readonly status: number;
  public readonly hanko: Hanko;
  public readonly invokedAction: string | undefined; // Changed from Set to single string
  public readonly autoStep?: TState extends AutoSteppedStates
    ? () => Promise<AnyState>
    : never;
  public readonly defaultHandler: TState extends DefaultHandledStates
    ? () => Promise<void>
    : never;

  // eslint-disable-next-line require-jsdoc
  constructor(hanko: Hanko, path: FlowPath, response: FlowResponse<TState>) {
    this.path = path;
    this.name = response.name;
    this.error = response.error;
    this.payload = response.payload;
    this.csrfToken = response.csrf_token;
    this.status = response.status;
    this.hanko = hanko;
    this.actions = this.buildActions(response.actions);

    if (this.name in autoSteps) {
      const handler = autoSteps[this.name as AutoSteppedStates];
      (this.autoStep as () => Promise<AnyState>) = () => handler(this as any);
    }

    if (this.name in defaultHandlers) {
      const handler = defaultHandlers[this.name as DefaultHandledStates];
      (this.defaultHandler as () => Promise<void>) = () => handler(this as any);
    }
  }

  // eslint-disable-next-line require-jsdoc
  private buildActions(actions: Actions[TState]): ActionMap<TState> {
    const actionMap: Partial<ActionMap<TState>> = {};

    Object.keys(actions).forEach((actionName) => {
      const key = actionName as keyof Actions[TState];
      const action = actions[key] as ActionType<any>;

      actionMap[key] = new Action(
        action,
        this.path,
        this.name,
        this.csrfToken,
        this.hanko,
        State.fetchState,
      );
    });

    // Return a Proxy that handles missing keys
    return new Proxy(actionMap as ActionMap<TState>, {
      get: (target: ActionMap<TState>, prop: string | symbol): Action<any> => {
        if (prop in target) {
          return target[prop as keyof ActionMap<TState>];
        }

        const actionName = typeof prop === "string" ? prop : prop.toString();

        return Action.createDisabled(
          actionName,
          this.path,
          this.name,
          this.csrfToken,
          this.hanko,
          State.fetchState,
        );
      },
    });
  }

  // eslint-disable-next-line require-jsdoc
  public dispatchEvents() {
    this.hanko.relay.dispatchFlowStateChangedEvent({ state: this as AnyState });
  }

  public hasAnyActionBeenInvoked(): boolean {
    return this.invokedAction !== undefined;
  }

  public recordActionInvocation(actionName: string): void {
    this.invokedAction = actionName;
  }

  public getInvokedAction(): string | undefined {
    return this.invokedAction;
  }

  // eslint-disable-next-line require-jsdoc
  public save(key: string): void {
    const serializedState: SerializedState = {
      path: this.path,
      name: this.name,
      error: this.error,
      payload: this.payload,
      csrf_token: this.csrfToken,
      status: this.status,
      actions:
        this.actions instanceof Proxy
          ? Object.fromEntries(Object.entries(this.actions))
          : this.actions, // Convert Proxy to plain object if needed
    };

    localStorage.setItem(key, JSON.stringify(serializedState));
  }

  // eslint-disable-next-line require-jsdoc
  public static load(hanko: Hanko, key: string): AnyState | null {
    const storedData = localStorage.getItem(key);
    if (!storedData) {
      return null;
    }
    const serializedState: SerializedState = JSON.parse(storedData);
    return new State(hanko, serializedState.path, serializedState);
  }

  // eslint-disable-next-line require-jsdoc
  public static async create(hanko: Hanko, path: FlowPath): Promise<AnyState> {
    const state = await State.fetchState(hanko, path, path);
    state.dispatchEvents();
    return state;
  }

  // eslint-disable-next-line require-jsdoc
  private static async fetchState(
    hanko: Hanko,
    path: FlowPath,
    href: string,
    body?: any,
  ): Promise<AnyState> {
    const response = await hanko.client.post(href, body);
    return new State(hanko, path, response.json()) as AnyState;
  }

  // eslint-disable-next-line require-jsdoc
  public static isState<T extends StateName>(
    state: State,
    name: T,
  ): state is State<T> {
    return state.name === name;
  }
}
