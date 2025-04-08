import { Hanko } from "../../Hanko";
import { Actions, Payloads, StateName } from "./types/state";
import { Input } from "./types/input";
import { Error } from "./types/error";
import { Action as ActionType } from "./types/action";
import { AnyState, FlowName, FlowResponse } from "./types/flow";
import { autoSteps } from "./auto-steps";
import { passkeyAutofillActivationHandlers } from "./passkey-autofill-activation";

export type AutoSteppedStates = keyof typeof autoSteps;
export type PasskeyAutofillStates =
  keyof typeof passkeyAutofillActivationHandlers;
export type AutoStepExclusion = AutoSteppedStates[] | "all";

export type ActionMap<TState extends StateName> = {
  [K in keyof Actions[TState]]: Action<
    Actions[TState][K] extends ActionType<infer TInputs> ? TInputs : never
  >;
};

export type ActionInfo = {
  name: string;
  relatedStateName: StateName;
};

export interface Options {
  dispatchAfterStateChangeEvent?: boolean;
  excludeAutoSteps?: AutoStepExclusion;
  previousAction?: ActionInfo;
  readFromLocalStorage?: boolean;
}

type SerializedState = FlowResponse<any> & {
  flow_name: FlowName;
  previous_action?: ActionInfo;
};

type ExtractInputValues<TInputs> = {
  [K in keyof TInputs]: TInputs[K] extends Input<infer TValue> ? TValue : never;
};

/**
 * Represents a state in a flow with associated actions and properties.
 * @template TState - The specific state name type.
 * @constructor
 * @param hanko - The Hanko instance for API interactions.
 * @param flowName - The name of the flow this state belongs to.
 * @param response - The flow response containing state data.
 * @param options - Configuration options for state initialization.
 * @category SDK
 * @subcategory FlowAPI
 */
export class State<TState extends StateName = StateName> {
  public readonly name: TState;
  public readonly flowName: FlowName;
  public error?: Error;
  public readonly payload?: Payloads[TState];
  public readonly actions: ActionMap<TState>;
  public readonly csrfToken: string;
  public readonly status: number;
  public readonly previousAction?: ActionInfo;
  public readonly readFromLocalStorage: boolean;
  public readonly hanko: Hanko;
  public invokedAction?: ActionInfo;
  public readonly excludeAutoSteps: AutoStepExclusion;

  public readonly autoStep?: TState extends AutoSteppedStates
    ? () => Promise<AnyState>
    : never;
  public readonly passkeyAutofillActivation: TState extends PasskeyAutofillStates
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
    this.actions = this.buildActionMap(response.actions);

    if (this.name in autoSteps) {
      const handler = autoSteps[this.name as AutoSteppedStates];
      (this.autoStep as () => Promise<AnyState>) = () => handler(this as any);
    }

    if (this.name in passkeyAutofillActivationHandlers) {
      const handler =
        passkeyAutofillActivationHandlers[this.name as PasskeyAutofillStates];
      (this.passkeyAutofillActivation as () => Promise<void>) = () =>
        handler(this as any);
    }

    const {
      dispatchAfterStateChangeEvent = true,
      excludeAutoSteps = null,
      previousAction = null,
      readFromLocalStorage = false,
    } = options;

    this.excludeAutoSteps = excludeAutoSteps;
    this.previousAction = previousAction;
    this.readFromLocalStorage = readFromLocalStorage;

    if (dispatchAfterStateChangeEvent) {
      this.dispatchAfterStateChangeEvent();
    }
  }

  /**
   * Builds the action map for this state.
   * @param actions - The actions available in this state.
   * @returns The action map for the state.
   * @private
   */
  private buildActionMap(actions: Actions[TState]): ActionMap<TState> {
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

  /**
   * Dispatches an event after the state has changed.
   */
  public dispatchAfterStateChangeEvent() {
    this.hanko.relay.dispatchAfterStateChangeEvent({
      state: this as AnyState,
    });
  }

  /**
   * Generates a local storage key for the given flow name.
   * @param flowName - The name of the flow.
   * @returns The formatted local storage key.
   */
  public static getLocalStorageKey(flowName: FlowName) {
    return `hanko_${flowName}_state`;
  }

  /**
   * Saves the current state to localStorage.
   */
  public saveToLocalStorage(): void {
    const serializedState: SerializedState = {
      flow_name: this.flowName,
      name: this.name,
      error: this.error,
      payload: this.payload,
      csrf_token: this.csrfToken,
      status: this.status,
      previous_action: this.previousAction,
      actions: Object.fromEntries(
        (Object.entries(this.actions) as [string, Action<any>][]).map(
          ([name, action]) => [
            name,
            {
              action: action.name,
              href: action.href,
              inputs: action.inputs,
            },
          ],
        ),
      ),
    };

    localStorage.setItem(
      State.getLocalStorageKey(this.flowName),
      JSON.stringify(serializedState),
    );
  }

  /**
   * Removes the current state from localStorage.
   */
  public removeFromLocalStorage() {
    localStorage.removeItem(State.getLocalStorageKey(this.flowName));
  }

  /**
   * Retrieves a flow response from localStorage.
   * @param flowName - The name of the flow.
   * @returns The stored flow state or null if not found.
   * @private
   */
  private static getFromLocalStorage(flowName: FlowName): SerializedState {
    const storedData = localStorage.getItem(State.getLocalStorageKey(flowName));
    if (!storedData) {
      return null;
    }
    return JSON.parse(storedData);
  }

  /**
   * Initializes a flow state, processing auto-steps if applicable.
   * @param hanko - The Hanko instance for API interactions.
   * @param flowName - The name of the flow.
   * @param response - The initial flow response.
   * @param options - Configuration options.
   * @returns A promise resolving to the initialized state.
   */
  public static async initializeFlowState(
    hanko: Hanko,
    flowName: FlowName,
    response: FlowResponse<any>,
    options: Options = {},
  ): Promise<AnyState> {
    let state = new State(hanko, flowName, response, options);

    if (state.excludeAutoSteps != "all") {
      while (
        state &&
        state.autoStep &&
        !state.excludeAutoSteps?.includes(state.name)
      ) {
        const nextState = await state.autoStep();
        if (nextState.name != state.name) {
          state = nextState;
        } else {
          return nextState;
        }
      }
    }

    return state;
  }

  /**
   * Creates a new state instance, using cached or fetched data.
   * @param hanko - The Hanko instance for API interactions.
   * @param flowName - The name of the flow.
   * @param options - Configuration options.
   * @returns A promise resolving to the created state.
   */
  public static async create(
    hanko: Hanko,
    flowName: FlowName,
    options: Omit<Options, "previousAction" | "readFromLocalStorage"> = {},
  ): Promise<AnyState> {
    const cachedState = State.getFromLocalStorage(flowName);

    if (cachedState) {
      return State.initializeFlowState(
        hanko,
        cachedState.flow_name,
        cachedState,
        {
          ...options,
          previousAction: cachedState.previous_action,
          readFromLocalStorage: true,
        },
      );
    }

    const newState = await State.fetchState(hanko, `/${flowName}`);

    return State.initializeFlowState(hanko, flowName, newState, options);
  }

  /**
   * Fetches state data from the server.
   * @param hanko - The Hanko instance for API interactions.
   * @param href - The endpoint to fetch from.
   * @param body - Optional request body.
   * @returns A promise resolving to the flow response.
   */
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

  /**
   * Creates an error flow response.
   * @param error - The error to include in the response.
   * @returns A flow response with error details.
   * @private
   */
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
}

/**
 * Represents an actionable operation within a state.
 * @template TInputs - The type of inputs required for the action.
 * @param action - The action type definition.
 * @param parentState - The state this action belongs to.
 * @param enabled - Whether the action is enabled (default: true).
 * @category SDK
 * @subcategory FlowAPI
 */
export class Action<TInputs> {
  public readonly enabled: boolean;
  public readonly href: string;
  public readonly name: string;
  public readonly inputs: TInputs;
  private readonly parentState: State;

  // eslint-disable-next-line require-jsdoc
  constructor(
    action: ActionType<TInputs>,
    parentState: State,
    enabled: boolean = true,
  ) {
    this.enabled = enabled;
    this.href = action.href;
    this.name = action.action;
    this.inputs = action.inputs;
    this.parentState = parentState;
  }

  /**
   * Creates a disabled action instance.
   * @param name - The name of the action.
   * @param parentState - The state this action belongs to.
   * @returns A disabled action instance.
   * @template TInputs - The type of inputs (inferred as empty).
   */
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

  /**
   * Executes the action, transitioning to a new state.
   * @param inputValues - Values for the action's inputs (optional).
   * @param options - Configuration options for execution.
   * @returns A promise resolving to the next state.
   * @throws Error if the action is disabled or already invoked.
   */
  async run(
    inputValues: ExtractInputValues<TInputs> = null,
    options: Pick<Options, "dispatchAfterStateChangeEvent"> = {},
  ): Promise<AnyState> {
    const {
      name,
      hanko,
      flowName,
      csrfToken,
      invokedAction,
      excludeAutoSteps,
    } = this.parentState;
    const { dispatchAfterStateChangeEvent = true } = options;

    if (!this.enabled) {
      throw new Error(
        `Action '${this.name}' is not enabled in state '${name}'`,
      );
    }

    if (invokedAction) {
      throw new Error(
        `An action '${invokedAction.name}' has already been invoked on state '${invokedAction.relatedStateName}'. No further actions can be run.`,
      );
    }

    this.parentState.invokedAction = {
      name: this.name,
      relatedStateName: name,
    };

    hanko.relay.dispatchBeforeStateChangeEvent({
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

    this.parentState.removeFromLocalStorage();

    return State.initializeFlowState(hanko, flowName, response, {
      dispatchAfterStateChangeEvent,
      excludeAutoSteps,
      previousAction: this.parentState.invokedAction,
    });
  }
}
