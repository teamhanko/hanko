import { Error } from "./error";
import { StateName, MappedPayloads, MappedActions } from "./state-handling";

export interface StateResponse<TStateName extends StateName> {
  name: StateName;
  status: number;
  payload?: MappedPayloads[TStateName];
  actions?: MappedActions[TStateName];
  error: Error;
}
