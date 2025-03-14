import { Actions, Payloads, StateName } from "./types/state";
import { Error } from "./types/error";
import { ActionType } from "./types/action-type";
import {
  ActionMap,
  AnyState,
  ExtractInputValues,
  FlowName,
  FlowResponse,
} from "./types/flow";
import { autoSteps, defaultHandlers } from "./auto-steps";
import { Hanko } from "../../Hanko";
import { Input } from "./types/input";

type AutoSteppedStates = keyof typeof autoSteps;
type DefaultHandledStates = keyof typeof defaultHandlers;
type SerializedState = FlowResponse<any> & { flowName: FlowName };

export interface Options {
  dispatchEvents?: boolean;
  runAutoSteps?: boolean;
}

// eslint-disable-next-line require-jsdoc
export class State<TState extends StateName = StateName> {
  public readonly name: TState;
  public readonly flowName: FlowName;
  public error?: Error;
  public readonly payload?: Payloads[TState];
  public readonly actions: ActionMap<TState>;
  public readonly csrfToken: string;
  public readonly status: number;

  public readonly hanko: Hanko;
  public invokedAction: string | undefined;
  public readonly runAutoSteps: boolean;

  public readonly autoStep?: TState extends AutoSteppedStates
    ? () => Promise<AnyState>
    : never;
  public readonly defaultHandler: TState extends DefaultHandledStates
    ? () => Promise<void>
    : never;

  // eslint-disable-next-line require-jsdoc
  constructor(
    hanko: Hanko,
    flowName: FlowName,
    response: FlowResponse<TState>,
    options: Options = {},
  ) {
    this.flowName = flowName;
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

    const { dispatchEvents = true, runAutoSteps = true } = options;

    this.runAutoSteps = runAutoSteps;

    if (dispatchEvents) {
      this.dispatchEvents();
    }
  }

  // eslint-disable-next-line require-jsdoc
  private buildActions(actions: Actions[TState]): ActionMap<TState> {
    const actionMap: Partial<ActionMap<TState>> = {};

    Object.keys(actions).forEach((actionName) => {
      const key = actionName as keyof Actions[TState];
      const action = actions[key] as ActionType<any>;

      actionMap[key] = new Action(action, this);
    });

    // Return a Proxy that handles missing keys
    return new Proxy(actionMap as ActionMap<TState>, {
      get: (target: ActionMap<TState>, prop: string | symbol): Action<any> => {
        if (prop in target) {
          return target[prop as keyof ActionMap<TState>];
        }

        const actionName = typeof prop === "string" ? prop : prop.toString();

        return Action.createDisabled(actionName, this);
      },
    });
  }

  // eslint-disable-next-line require-jsdoc
  public dispatchEvents() {
    this.hanko.relay.dispatchFlowStateChangedEvent({ state: this as AnyState });
  }

  // eslint-disable-next-line require-jsdoc
  public save(key: string): void {
    const serializedState: SerializedState = {
      flowName: this.flowName,
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
    return new State(hanko, serializedState.flowName, serializedState);
  }

  private static async initializeFlowState(
    hanko: Hanko,
    flowName: FlowName,
    response: FlowResponse<any>,
    options: Options = {},
  ): Promise<AnyState> {
    let state = new State(hanko, flowName, response, options);

    while (state.autoStep) {
      state = await state.autoStep();
    }

    return state;
  }

  // eslint-disable-next-line require-jsdoc
  public static async init(
    hanko: Hanko,
    flowName: FlowName,
  ): Promise<AnyState> {
    const response = await State.fetchState(hanko, `/${flowName}`);
    return new State(hanko, flowName, response) as AnyState;
  }

  // eslint-disable-next-line require-jsdoc
  static async fetchState(
    hanko: Hanko,
    href: string,
    body?: any,
  ): Promise<FlowResponse<any>> {
    try {
      const response = await hanko.client.post(href, body);
      return response.json();
    } catch (error) {
      return State.createErrorResponse(error);
    }
  }

  // eslint-disable-next-line require-jsdoc
  private static createErrorResponse(error: Error): FlowResponse<"error"> {
    return {
      actions: null,
      csrf_token: "",
      name: "error",
      payload: null,
      status: 0,
      error,
    };
  }

  // eslint-disable-next-line require-jsdoc
  public static isState<T extends StateName>(
    state: State,
    name: T,
  ): state is State<T> {
    return state.name === name;
  }
}

// eslint-disable-next-line require-jsdoc
export class Action<TInputs> {
  private readonly href: string;
  private readonly parentState: State;
  public readonly name: string;
  public readonly enabled: boolean;
  public readonly inputs: TInputs;

  // eslint-disable-next-line require-jsdoc
  constructor(
    action: ActionType<TInputs>,
    parentState: State,
    enabled: boolean = true,
  ) {
    this.enabled = enabled;
    this.inputs = action.inputs;
    this.href = action.href;
    this.name = action.action;
    this.parentState = parentState;
  }

  // eslint-disable-next-line require-jsdoc
  static createDisabled<TInputs>(
    name: string,
    parentState: State,
  ): Action<TInputs> {
    return new Action(
      {
        action: name,
        href: "", // No valid href since it’s disabled
        inputs: {} as TInputs,
        description: "Disabled action",
      },
      parentState,
      false,
    );
  }

  // eslint-disable-next-line require-jsdoc
  async run(
    inputValues: ExtractInputValues<TInputs> = null,
    runOptions: Options = {},
  ): Promise<AnyState> {
    const { dispatchEvents = true } = runOptions;
    const { name, hanko, flowName, csrfToken, invokedAction } =
      this.parentState;

    if (!this.enabled) {
      throw new Error(
        `Action '${this.name}' is not enabled in state '${name}'`,
      );
    }

    if (invokedAction) {
      throw new Error(
        `An action '${invokedAction}' has already been invoked on state '${name}'. No further actions can be run.`,
      );
    }

    this.parentState.invokedAction = this.name;

    hanko.relay.dispatchFlowBeforeStateChangedEvent({
      state: this.parentState as AnyState,
    });

    // Extract default values from this.inputs
    const defaultValues = Object.keys(this.inputs).reduce(
      (acc, key) => {
        const input = (this.inputs as any)[key] as Input<any>;
        if (input.value !== undefined) {
          acc[key] = input.value;
        }
        return acc;
      },
      {} as Record<string, any>,
    );

    // Merge defaults with user-provided inputs
    const mergedInputData = {
      ...defaultValues,
      ...inputValues,
    };

    const requestBody = {
      input_data: mergedInputData,
      csrf_token: csrfToken,
    };

    const response = await State.fetchState(hanko, this.href, requestBody);

    return new State(hanko, flowName, response, { dispatchEvents }) as AnyState;
  }
}
