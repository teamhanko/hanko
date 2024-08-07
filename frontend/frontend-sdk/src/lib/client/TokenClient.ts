import { Client } from "./Client";
import { TechnicalError } from "../Errors";

/**
 * Client responsible for exchanging one time tokens for session JWTs.
 *
 * @constructor
 * @category SDK
 * @subcategory Clients
 * @extends {Client}
 */
export class TokenClient extends Client {
  /**
   * Validate a one time token to retrieve a session JWT. Does nothing
   * if the current window location does not contain a 'hanko_token' in the
   * search query.
   *
   * @return {Promise<void>}
   * @throws {TechnicalError}
   * https://docs.hanko.io/api/api/public#tag/Token/operation/token
   */
  async validate(): Promise<void> {
    const params = new URLSearchParams(window.location.search);
    const token = params.get("hanko_token");

    if (!token) return;

    window.history.replaceState(null, null, window.location.pathname);

    const response = await this.client.post("/token", { value: token });
    if (!response.ok) {
      throw new TechnicalError();
    }

    return response.json();
  }
}
