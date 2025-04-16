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

export interface AllOptions {
  dispatchAfterStateChangeEvent?: boolean;
  excludeAutoSteps?: AutoStepExclusion;
  previousAction?: ActionInfo;
  isCached?: boolean;
  cacheKey?: string;
}

export type Options = Omit<AllOptions, "previousAction" | "isCached"> & {
  loadFromCache?: boolean;
};

type SerializedState = FlowResponse<any> & {
  flow_name: FlowName;
  previous_action?: ActionInfo;
  is_cached?: boolean;
};

type ExtractInputValues<TInputs> = {
  [K in keyof TInputs]: TInputs[K] extends Input<infer TValue> ? TValue : never;
};

/**
 * Represents a state in a flow with associated actions and properties.
 * @template TState - The specific state name type.
 * @constructor
 * @param {Hanko} hanko - The Hanko instance for API interactions.
 * @param {FlowName} flowName - The name of the flow this state belongs to.
 * @param {FlowResponse<TState>} response - The flow response containing state data.
 * @param {AllOptions} [options={}] - Configuration options for state initialization.
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
  public readonly isCached: boolean;
  public readonly cacheKey: string;
  public readonly hanko: Hanko;
  public invokedAction?: ActionInfo;
  public readonly excludeAutoSteps: AutoStepExclusion;

  public readonly autoStep?: TState extends AutoSteppedStates
    ? () => Promise<AnyState>
    : never;
  public readonly passkeyAutofillActivation: TState extends PasskeyAutofillStates
    ? () => Promise<void>
    : never;

  /**
   * Constructs a new State instance.
   * @param {Hanko} hanko - The Hanko instance for API interactions.
   * @param {FlowName} flowName - The name of the flow this state belongs to.
   * @param {FlowResponse<TState>} response - The flow response containing state data.
   * @param {AllOptions} [options={}] - Configuration options for state initialization.
   */
  constructor(
    hanko: Hanko,
    flowName: FlowName,
    response: FlowResponse<TState>,
    options: AllOptions = {},
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
      isCached = false,
      cacheKey = "hanko-flow-state",
    } = options;

    this.excludeAutoSteps = excludeAutoSteps;
    this.previousAction = previousAction;
    this.isCached = isCached;
    this.cacheKey = cacheKey;

    if (dispatchAfterStateChangeEvent) {
      this.dispatchAfterStateChangeEvent();
    }
  }

  /**
   * Builds the action map for this state, wrapping it in a Proxy to handle undefined actions.
   * @param {Actions} actions - The actions available in this state.
   * @returns {ActionMap<TState>} The action map for the state.
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
   * Serializes the current state into a storable format.
   * @returns {SerializedState} The serialized state object.
   */
  public serialize(): SerializedState {
    return {
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
              description: null,
            },
          ],
        ),
      ),
    };
  }

  /**
   * Saves the current state to localStorage.
   * @returns {void}
   */
  public saveToLocalStorage(): void {
    localStorage.setItem(
      this.cacheKey,
      JSON.stringify({ ...this.serialize(), is_cached: true }),
    );
  }

  /**
   * Removes the current state from localStorage.
   * @returns {void}
   */
  public removeFromLocalStorage(): void {
    localStorage.removeItem(this.cacheKey);
  }

  /**
   * Initializes a flow state, processing auto-steps if applicable.
   * @param {Hanko} hanko - The Hanko instance for API interactions.
   * @param {FlowName} flowName - The name of the flow.
   * @param {FlowResponse<any>} response - The initial flow response.
   * @param {AllOptions} [options={}] - Configuration options.
   * @param {boolean} [options.dispatchAfterStateChangeEvent=true] - Whether to dispatch an event after state change.
   * @param {AutoStepExclusion} [options.excludeAutoSteps=null] - States to exclude from auto-step processing, or "all".
   * @param {ActionInfo} [options.previousAction=null] - Information about the previous action.
   * @param {boolean} [options.isCached=false] - Whether the state is loaded from cache.
   * @param {string} [options.cacheKey="hanko-flow-state"] - Key for localStorage caching.
   * @returns {Promise<AnyState>} A promise resolving to the initialized state.
   */
  public static async initializeFlowState(
    hanko: Hanko,
    flowName: FlowName,
    response: FlowResponse<any>,
    options: AllOptions = {},
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
   * Retrieves and parses state data from localStorage.
   * @param {string} cacheKey - The key used to store the state in localStorage.
   * @returns {SerializedState | undefined} The parsed serialized state, or undefined if not found or invalid.
   */
  public static readFromLocalStorage(
    cacheKey: string,
  ): SerializedState | undefined {
    const raw = localStorage.getItem(cacheKey);
    if (raw) {
      try {
        return JSON.parse(raw) as SerializedState;
      } catch {
        return undefined;
      }
    }
  }

  /**
   * Creates a new state instance, using cached or fetched data.
   * @param {Hanko} hanko - The Hanko instance for API interactions.
   * @param {FlowName} flowName - The name of the flow.
   * @param {Options} [options={}] - Configuration options.
   * @param {boolean} [options.dispatchAfterStateChangeEvent=true] - Whether to dispatch an event after state change.
   * @param {AutoStepExclusion} [options.excludeAutoSteps=null] - States to exclude from auto-step processing, or "all".
   * @param {string} [options.cacheKey="hanko-flow-state"] - Key for localStorage caching.
   * @param {boolean} [options.loadFromCache=true] - Whether to attempt loading from cache.
   * @returns {Promise<AnyState>} A promise resolving to the created state.
   */
  public static async create(
    hanko: Hanko,
    flowName: FlowName,
    options: Options = {},
  ): Promise<AnyState> {
    const { cacheKey = "hanko-flow-state", loadFromCache = true } = options;
    if (loadFromCache) {
      const cachedState = State.readFromLocalStorage(cacheKey);
      if (cachedState) {
        return State.deserialize(hanko, cachedState, {
          ...options,
          cacheKey,
        });
      }
    }

    const newState = await State.fetchState(hanko, `/${flowName}`);
    return State.initializeFlowState(hanko, flowName, newState, {
      ...options,
      cacheKey,
    });
  }

  /**
   * Deserializes a state from a serialized state object.
   * @param {Hanko} hanko - The Hanko instance for API interactions.
   * @param {SerializedState} serializedState - The serialized state data.
   * @param {Options} [options={}] - Configuration options.
   * @param {boolean} [options.dispatchAfterStateChangeEvent=true] - Whether to dispatch an event after state change.
   * @param {AutoStepExclusion} [options.excludeAutoSteps=null] - States to exclude from auto-step processing, or "all".
   * @param {string} [options.cacheKey="hanko-flow-state"] - Key for localStorage caching.
   * @param {boolean} [options.loadFromCache=true] - Whether to attempt loading from cache.
   * @returns {Promise<AnyState>} A promise resolving to the deserialized state.
   */
  public static async deserialize(
    hanko: Hanko,
    serializedState: SerializedState,
    options: Options = {},
  ): Promise<AnyState> {
    return State.initializeFlowState(
      hanko,
      serializedState.flow_name,
      serializedState,
      {
        ...options,
        previousAction: serializedState.previous_action,
        isCached: serializedState.is_cached,
      },
    );
  }

  /**
   * Fetches state data from the server.
   * @param {Hanko} hanko - The Hanko instance for API interactions.
   * @param {string} href - The endpoint to fetch from.
   * @param {any} [body] - Optional request body.
   * @returns {Promise<FlowResponse<any>>} A promise resolving to the flow response.
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
   * @param {Error} error - The error to include in the response.
   * @returns {FlowResponse<"error">} A flow response with error details.
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
 * @param {ActionType<TInputs>} action - The action type definition.
 * @param {State} parentState - The state this action belongs to.
 * @param {boolean} [enabled=true] - Whether the action is enabled.
 * @category SDK
 * @subcategory FlowAPI
 */
export class Action<TInputs> {
  public readonly enabled: boolean;
  public readonly href: string;
  public readonly name: string;
  public readonly inputs: TInputs;
  private readonly parentState: State;

  /**
   * Constructs a new Action instance.
   * @param {ActionType<TInputs>} action - The action type definition.
   * @param {State} parentState - The state this action belongs to.
   * @param {boolean} [enabled=true] - Whether the action is enabled.
   */
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
   * @template TInputs - The type of inputs (inferred as empty).
   * @param {string} name - The name of the action.
   * @param {State} parentState - The state this action belongs to.
   * @returns {Action<TInputs>} A disabled action instance.
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
   * @param {ExtractInputValues<TInputs>} [inputValues=null] - Values for the action's inputs.
   * @param {Pick<AllOptions, "dispatchAfterStateChangeEvent">} [options={}] - Configuration options.
   * @param {boolean} [options.dispatchAfterStateChangeEvent=true] - Whether to dispatch an event after state change.
   * @returns {Promise<AnyState>} A promise resolving to the next state.
   * @throws {Error} If the action is disabled or already invoked.
   */
  async run(
    inputValues: ExtractInputValues<TInputs> = null,
    options: Pick<AllOptions, "dispatchAfterStateChangeEvent"> = {},
  ): Promise<AnyState> {
    const {
      name,
      hanko,
      flowName,
      csrfToken,
      invokedAction,
      excludeAutoSteps,
      cacheKey,
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
      previousAction: invokedAction,
      cacheKey,
    });
  }
}
