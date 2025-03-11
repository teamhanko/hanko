import { ActionType as ActionType } from "./types/actionType";
import { AnyState, ExtractInputValues, FetchFunction, FlowPath } from "./types/flow";
import { StateName } from "./types/state";
import { Input } from "./types/input";
import { Hanko } from "../../Hanko";
import { ActionState } from "./types/action-state";

export interface RunOptions {
  dispatchEvents?: boolean;
}

// eslint-disable-next-line require-jsdoc
export class Action<TInputs> {
  private readonly href: string;
  private readonly fetchState: FetchFunction;
  private readonly name: string;
  private readonly stateName: StateName;
  private readonly hanko: Hanko;
  private readonly path: FlowPath;
  private readonly csrfToken: string;
  private readonly parentState: ActionState;
  public readonly enabled: boolean;
  public readonly inputs: TInputs;

  // eslint-disable-next-line require-jsdoc
  constructor(
    action: ActionType<TInputs>,
    parentState: ActionState,
    enabled: boolean = true,
  ) {
    this.path = path;
    this.enabled = enabled;
    this.inputs = action.inputs;
    this.href = action.href;
    this.name = action.name;
    this.stateName = stateName;
    this.csrfToken = csrfToken;
    this.hanko = hanko;
    this.fetchState = fetchState;
  }

  static createDisabled<TInputs>(
    actionName: string,
    path: FlowPath,
    stateName: StateName,
    csrfToken: string,
    hanko: Hanko,
    fetchState: FetchFunction,
  ): Action<TInputs> {
    return new Action(
      {
        name: actionName,
        href: "", // No valid href since it’s disabled
        inputs: {} as TInputs,
        description: "Disabled action",
      },
      path,
      stateName,
      csrfToken,
      hanko,
      fetchState,
      false,
    );
  }

  // eslint-disable-next-line require-jsdoc
  async run(inputValues: ExtractInputValues<TInputs> = null, runOptions: RunOptions = {}): Promise<AnyState> {
    const { dispatchEvents = true } = runOptions;

    if (!this.enabled) {
      throw new Error(
        `Action '${this.name}' is not enabled in state '${this.stateName}'`,
      );
    }

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
      csrf_token: this.csrfToken,
    };

    const state = await this.fetchState(this.hanko, this.path, this.href, requestBody)

    if (dispatchEvents) {
      this.hanko.relay.dispatchFlowStateChangedEvent({state});
    }

    return state;
  }
}
