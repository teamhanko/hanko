import { Client } from "./Client";
import { SessionCheckResponse } from "../Dto";
import { TechnicalError } from "../Errors";

/**
 * A class that handles communication with the Hanko API for the purposes
 * of sessions.
 *
 * @constructor
 * @category SDK
 * @subcategory Clients
 * @extends {Client}
 */
export class SessionClient extends Client {
  /**
   * Checks if the current session is still valid.
   *
   * @return {Promise<SessionCheckResponse>}
   * @throws {TechnicalError}
   */
  async isValid(): Promise<SessionCheckResponse> {
    const response = await this.client.get("/sessions/check");

    if (!response.ok) {
      throw new TechnicalError();
    }

    return response.json();
  }
}
