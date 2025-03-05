import { ActionType as ActionType } from "./types/actionType";
import { AllStates, FetchFunction, ExtractInputValues } from "./types/flow";
import { StateName } from "./types/state";
import { Input } from "./types/input";

// eslint-disable-next-line require-jsdoc
export class Action<TInputs> {
  public readonly enabled: boolean;
  public readonly inputs: TInputs;
  private readonly href: string;
  private readonly fetchFunc: FetchFunction;
  private readonly name: string;
  private readonly stateName: StateName;
  private readonly csrfToken: string;

  // eslint-disable-next-line require-jsdoc
  constructor(
    action: ActionType<TInputs>,
    fetchFunc: FetchFunction,
    stateName: StateName,
    csrfToken: string,
    enabled: boolean = false,
  ) {
    this.enabled = enabled;
    this.inputs = action.inputs;
    this.href = action.href;
    this.fetchFunc = fetchFunc;
    this.name = action.name;
    this.stateName = stateName;
    this.csrfToken = csrfToken;
  }

  static createDisabled<TInputs>(
    stateName: StateName,
    fetchFunc: FetchFunction,
    csrfToken: string
  ): Action<TInputs> {
    return new Action(
      {
        name: "disabled",
        href: "", // No valid href since it’s disabled
        inputs: {} as TInputs,
        description: "Disabled action",
      },
      fetchFunc,
      stateName,
      csrfToken,
      false,
    );
  }

  // eslint-disable-next-line require-jsdoc
  async run(inputValues: ExtractInputValues<TInputs>): Promise<AllStates> {
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

    return this.fetchFunc(this.href, requestBody);
  }
}
