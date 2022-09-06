import { Config } from "../Dto";
import { TechnicalError } from "../Errors";
import { Client } from "./Client";

/**
 * A class for retrieving configurations from the API.
 *
 * @category SDK
 * @subcategory Clients
 * @extends {Client}
 */
class ConfigClient extends Client {
  /**
   * Retrieves the frontend configuration.
   * @return {Promise<Config>}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://teamhanko.github.io/hanko/#/.well-known/getConfig
   */
  get() {
    return new Promise<Config>((resolve, reject) => {
      this.client
        .get("/.well-known/config")
        .then((response) => {
          if (response.ok) {
            return resolve(response.json());
          }

          throw new TechnicalError();
        })
        .catch((e) => {
          reject(e);
        });
    });
  }
}

export { ConfigClient };
