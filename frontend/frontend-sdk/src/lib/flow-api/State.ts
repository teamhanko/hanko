import { Actions, Payloads, StateName } from "./types/state";
import { Error } from "./types/error";
import { ActionType } from "./types/actionType";
import {
  ActionMap,
  AllStates,
  FetchFunction,
  FlowResponse,
} from "./types/flow";
import { Action } from "./Action";
import { autoSteps } from "./auto-steps";

type AutoSteppedStates = keyof typeof autoSteps;
type ConditionalAutoStepGuard<TState> = TState extends AutoSteppedStates
  ? () => Promise<AllStates>
  : never;

// eslint-disable-next-line require-jsdoc
export class State<TState extends StateName = StateName> {
  public readonly name: TState;
  public readonly error?: Error;
  public readonly payload?: Payloads[TState];
  public readonly actions: ActionMap<TState>;
  private readonly csrfToken: string;
  private readonly fetchFunc: FetchFunction;
  public readonly autoStep?: ConditionalAutoStepGuard<TState>;

  // public readonly autoStep?: () => Promise<AllStates>;

  // eslint-disable-next-line require-jsdoc
  constructor(response: FlowResponse<TState>, fetchFunc: FetchFunction) {
    this.name = response.name; // No cast needed
    this.error = response.error;
    this.payload = response.payload;
    this.csrfToken = response.csrf_token;
    this.actions = this.buildActions(response.actions);
    this.fetchFunc = fetchFunc;
    this.autoStep = this.getAutoStep();
  }

  // eslint-disable-next-line require-jsdoc
  private buildActions(actions: Actions[TState]): ActionMap<TState> {
    const actionMap: Partial<ActionMap<TState>> = {};

    Object.keys(actions).forEach((actionName) => {
      const key = actionName as keyof Actions[TState];
      const action = actions[key] as ActionType<any>;
      actionMap[key] = new Action(
        action,
        this.fetchFunc,
        this.name,
        this.csrfToken,
      );
    });

    return actionMap as ActionMap<TState>;
  }

  // eslint-disable-next-line require-jsdoc
  private getAutoStep(): ConditionalAutoStepGuard<TState> {
    if (isAutoSteppedState(this.name)) {
      const handler = autoSteps[this.name];
      return (() => handler(this as any)) as ConditionalAutoStepGuard<TState>;
    }
    return;
  }
}

// eslint-disable-next-line require-jsdoc
function isAutoSteppedState(name: StateName): name is AutoSteppedStates {
  return name in autoSteps;
}
