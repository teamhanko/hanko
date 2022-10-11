import { Config } from "../Dto";
import { TechnicalError } from "../Errors";
import { Client } from "./Client";

/**
 * A class for retrieving configurations from the API.
 *
 * @category SDK
 * @subcategory Clsients
 * @extends {Client}
 */
class ConfigClient extends Client {
  /**
   * Retrieves the frontend configuration.
   * @return {Promise<Config>}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api/public#tag/.well-known/operation/getConfig
   */
  async get(): Promise<Config> {
    const response = await this.client.get("/.well-known/config");

    if (!response.ok) {
      throw new TechnicalError();
    }

    return response.json();
  }
}

export { ConfigClient };
