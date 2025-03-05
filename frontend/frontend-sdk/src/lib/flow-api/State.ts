import { Actions, Payloads, StateName } from "./types/state";
import { Error } from "./types/error";
import { ActionType } from "./types/actionType";
import {
  ActionMap,
  AllStates,
  FetchFunction,
  FlowResponse,
  DefaultHandlers,
} from "./types/flow";
import { Action } from "./Action";
import { autoSteps, defaultHandlers } from "./auto-steps";

type AutoSteppedStates = keyof typeof autoSteps;
type DefaultHandledStates = keyof typeof defaultHandlers;

// eslint-disable-next-line require-jsdoc
export class State<TState extends StateName = StateName> {
  public readonly name: TState;
  public error?: Error;
  public readonly payload?: Payloads[TState];
  public readonly actions: ActionMap<TState>;
  private readonly csrfToken: string;
  private readonly fetchFunc: FetchFunction;
  public readonly autoStep?: TState extends AutoSteppedStates
    ? () => Promise<AllStates>
    : never;
  public readonly defaultHandler: TState extends DefaultHandledStates
    ? () => Promise<void>
    : never;

  // eslint-disable-next-line require-jsdoc
  constructor(response: FlowResponse<TState>, fetchFunc: FetchFunction) {
    this.name = response.name;
    this.error = response.error;
    this.payload = response.payload;
    this.csrfToken = response.csrf_token;
    this.actions = this.buildActions(response.actions);
    this.fetchFunc = fetchFunc;

    if (this.name in autoSteps) {
      const handler = autoSteps[this.name as AutoSteppedStates];
      (this.autoStep as () => Promise<AllStates>) = () => handler(this as any);
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
        this.fetchFunc,
        this.name,
        this.csrfToken,
      );
    });

    return actionMap as ActionMap<TState>;
  }
}
