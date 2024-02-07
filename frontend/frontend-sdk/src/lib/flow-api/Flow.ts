import { Client } from "../client/Client";
import { State } from "./State";
import { FetchNextState, FlowPath } from "./types/state-handling";

// eslint-disable-next-line require-jsdoc
class Flow extends Client {
  public async init(path: FlowPath): Promise<State<any>> {
    const fetchState: FetchNextState = async (href: string, body?: any) => {
      const response = await this.client.post(href, body);
      return new State(response.json(), fetchState);
    };

    // Start the flow execution
    return fetchState(path);
  }
}

export { Flow };
