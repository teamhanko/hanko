import { FetchFunction, FlowPath } from "./flow";
import { StateName } from "./state";
import { Hanko } from "../../../Hanko";

export interface ActionState {
  // eslint-disable-next-line no-unused-vars
  recordActionInvocation(actionName: string): void;
  // eslint-disable-next-line no-unused-vars
  getInvokedAction(): string | undefined;
  name: StateName;
  fetchFunc: FetchFunction;
  csrfToken: string;
  hanko: Hanko;
  path: FlowPath;
}
