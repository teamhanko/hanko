import { TechnicalError, UnauthorizedError } from "../Errors";
import { Client } from "./Client";
import { User } from "../flow-api/types/payload";
import { Me } from "../Dto";

/**
 * A class to manage user information.
 *
 * @category SDK
 * @subcategory Clients
 * @extends {Client}
 */
class UserClient extends Client {
  /**
   * Fetches the current user.
   *
   * @return {Promise<User>}
   * @throws {UnauthorizedError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api/public#tag/User-Management/operation/IsUserAuthorized
   * @see https://docs.hanko.io/api/public#tag/User-Management/operation/listUser
   */
  async getCurrent(): Promise<User> {
    const meResponse = await this.client.get("/me");

    if (meResponse.status === 401) {
      this.client.dispatcher.dispatchSessionExpiredEvent();
      throw new UnauthorizedError();
    } else if (!meResponse.ok) {
      throw new TechnicalError();
    }

    const me: Me = meResponse.json();
    const userResponse = await this.client.get(`/users/${me.id}`);

    if (userResponse.status === 401) {
      this.client.dispatcher.dispatchSessionExpiredEvent();
      throw new UnauthorizedError();
    } else if (!userResponse.ok) {
      throw new TechnicalError();
    }

    return userResponse.json();
  }

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
