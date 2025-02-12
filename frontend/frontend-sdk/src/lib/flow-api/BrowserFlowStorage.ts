import { FlowOld } from "./FlowOld";
import { StateOld } from "./StateOld";

export default class BrowserFlowStorage {
  static save(state: StateOld<any>, key = "hanko-state") {
    localStorage.setItem(key, JSON.stringify(state));
  }

  static load(
    flow: FlowOld | FlowOld["fetchNextState"],
    key = "hanko-state"
  ): StateOld<any> | null {
    const state = localStorage.getItem(key);

    const fetchNextState =
      typeof flow === "function"
        ? flow // `flow` is already a fetchNextState function
        : flow.fetchNextState.bind(flow); // `flow` is a `Flow`

    if (state) {
      return new StateOld(JSON.parse(state), fetchNextState);
    }

    return null;
  }
}
