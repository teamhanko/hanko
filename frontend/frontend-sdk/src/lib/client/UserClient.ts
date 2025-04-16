import { TechnicalError } from "../Errors";
import { Client } from "./Client";

/**
 * A class to manage user information.
 *
 * @category SDK
 * @subcategory Clients
 * @extends {Client}
 */
class UserClient extends Client {
  /**
   * Logs out the current user and expires the existing session cookie. A valid session cookie is required to call the logout endpoint.
   *
   * @return {Promise<void>}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   */
  async logout(): Promise<void> {
    const logoutResponse = await this.client.post("/logout");

    // For cross-domain operations, the frontend SDK creates the cookie by reading the "X-Auth-Token" header, and
    // "Set-Cookie" headers sent by the backend have no effect due to the browser's security policy, which means that
    // the cookie must also be removed client-side in that case.
    this.client.cookie.removeAuthCookie();
    this.client.dispatcher.dispatchUserLoggedOutEvent();

    if (logoutResponse.status === 401) {
      // The user is logged out already
      return;
    } else if (!logoutResponse.ok) {
      throw new TechnicalError();
    }
  }
}

export { UserClient };
