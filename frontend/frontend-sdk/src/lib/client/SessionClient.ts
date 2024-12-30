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
  async validate(): Promise<SessionCheckResponse> {
    const response = await this.client.get("/sessions/validate");

    if (!response.ok) {
      throw new TechnicalError();
    }

    return await response.json();
  }
}

// Class to maintain compatibility with previous versions.
export class Session extends Client {
  /**
   * Checks if the current session is still valid. This function is to be removed - please replace
   * any usage with the new 'SessionClient.validate()' function.
   *
   * @return {boolean}
   * @throws {TechnicalError}
   * @deprecated
   */
  isValid(): boolean {
    let session: SessionCheckResponse;
    try {
      const response = this.client._fetch_blocking("/sessions/validate", {
        method: "GET",
      });
      session = JSON.parse(response);
    } catch (e) {
      throw new TechnicalError(e);
    }
    return session ? session.is_valid : false;
  }
}
