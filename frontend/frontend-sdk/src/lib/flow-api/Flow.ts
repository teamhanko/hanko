import { Client } from "../client/Client";
import { FlowPath } from "./types/state-handling";
import { State } from "./State";

// eslint-disable-next-line require-jsdoc
class Flow extends Client {
  // eslint-disable-next-line require-jsdoc
  public async init(initPath: FlowPath): Promise<State<any>> {
    const fetchState = async (href: string, body?: any) => {
      const response = await this.client.post(href, body);
      return new State(response.json(), fetchState);
    };

    return fetchState(initPath);
  }
}

export { Flow };
