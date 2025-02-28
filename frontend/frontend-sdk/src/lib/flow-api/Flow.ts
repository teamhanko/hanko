import { Client } from "../client/Client";
import { FlowPath, StateName } from "./types/state-handling";
import { State, FetchFunction, AllStates } from "./State";

// eslint-disable-next-line require-jsdoc
class Flow extends Client {
  // eslint-disable-next-line require-jsdoc
  async init(initPath: FlowPath): Promise<AllStates> {
    const fetchState: FetchFunction = async (href: string, body?: any) => {
      const response = await this.client.post(href, body);
      return new State(await response.json(), fetchState) as AllStates;
    };
    return fetchState(initPath);
  }

  // eslint-disable-next-line require-jsdoc
  static isState<T extends StateName>(
    state: State,
    name: T,
  ): state is State<T> {
    return state.name === name;
  }
}

export { Flow };
