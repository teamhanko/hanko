import { Flow } from "./Flow";
import { State } from "./State";

export default class BrowserFlowStorage {
  static save(state: State<any>, key = "hanko-state") {
    localStorage.setItem(key, JSON.stringify(state));
  }

  static load(
    flow: Flow | Flow["fetchNextState"],
    key = "hanko-state"
  ): State<any> | null {
    const state = localStorage.getItem(key);

    const fetchNextState =
      typeof flow === "function"
        ? flow // `flow` is already a fetchNextState function
        : flow.fetchNextState.bind(flow); // `flow` is a `Flow`

    if (state) {
      return new State(JSON.parse(state), fetchNextState);
    }

    return null;
  }
}
