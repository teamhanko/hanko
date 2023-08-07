import { Me, User, UserInfo, UserCreated } from "../Dto";
import {
  ConflictError,
  NotFoundError,
  TechnicalError,
  UnauthorizedError,
  ForbiddenError,
} from "../Errors";
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
   * Fetches basic information about the user identified by the given email address. Can be used while the user is logged out
   * and is helpful in deciding which type of login to choose. For example, if the user's email is not verified, you may
   * want to log in with a passcode, or if no WebAuthn credentials are registered, you may not want to use WebAuthn.
   *
   * @param {string} email - The user's email address.
   * @return {Promise<UserInfo>}
   * @throws {NotFoundError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api/public#tag/User-Management/operation/getUserId
   */
  async getInfo(email: string): Promise<UserInfo> {
    const response = await this.client.post("/user", { email });

    if (response.status === 404) {
      throw new NotFoundError();
    } else if (!response.ok) {
      throw new TechnicalError();
    }

    return response.json();
  }

  /**
   * Creates a new user. Afterwards, verify the email address via passcode. If a 'ConflictError'
   * occurred, you may want to prompt the user to log in.
   *
   * @param {string} email - The email address of the user to be created.
   * @return {Promise<UserCreated>}
   * @throws {ConflictError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api/public#tag/User-Management/operation/createUser
   */
  async create(email: string): Promise<UserCreated> {
    const response = await this.client.post("/users", { email });

    if (response.status === 409) {
      throw new ConflictError();
    } if (response.status === 403) {
      throw new ForbiddenError();
    } else if (!response.ok) {
      throw new TechnicalError();
    }

    const createUser: UserCreated = response.json();
    if (createUser && createUser.user_id) {
      this.client.processResponseHeadersOnLogin(createUser.user_id, response);
    }
    return createUser;
  }

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
   * Deletes the current user and expires the existing session cookie.
   *
   * @return {Promise<void>}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @throws {UnauthorizedError}
   */
  async delete(): Promise<void> {
    const response = await this.client.delete("/user");

    if (response.ok) {
      this.client.cookie.removeAuthCookie();
      this.client.sessionState.reset().write();
      this.client.dispatcher.dispatchUserDeletedEvent();
      return;
    } else if (response.status === 401) {
      this.client.dispatcher.dispatchSessionExpiredEvent();
      throw new UnauthorizedError();
    }

    throw new TechnicalError();
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
    this.client.sessionState.reset().write();
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
