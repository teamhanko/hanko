import { Me, User, UserInfo } from "../Dto";
import {
  ConflictError,
  NotFoundError,
  TechnicalError,
  UnauthorizedError,
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
   * Fetches basic information about the user by providing an email address. Can be used while the user is logged out
   * and is helpful in deciding which type of login to choose. For example, if the user's email is not verified, you may
   * want to log in with a passcode, or if no WebAuthN credentials are registered, you may not want to use WebAuthN.
   *
   * @param {string} email - The user's email address.
   * @return {Promise<UserInfo>}
   * @throws {NotFoundError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://teamhanko.github.io/hanko/#/authentication/getUserId
   */
  getInfo(email: string): Promise<UserInfo> {
    return new Promise<UserInfo>((resolve, reject) => {
      this.client
        .post("/user", { email })
        .then((response) => {
          if (response.ok) {
            return response.json();
          } else if (response.status === 404) {
            throw new NotFoundError();
          } else {
            throw new TechnicalError();
          }
        })
        .then((u: UserInfo) => resolve(u))
        .catch((e) => {
          reject(e);
        });
    });
  }

  /**
   * Creates a new user. Afterwards, verify the email address via passcode. If a 'ConflictError'
   * occurred, you may want to prompt the user to log in.
   *
   * @param {string} email - The email address of the user to be created.
   * @return {Promise<User>}
   * @throws {ConflictError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://teamhanko.github.io/hanko/#/management/createUser
   */
  create(email: string): Promise<User> {
    return new Promise<User>((resolve, reject) => {
      this.client
        .post("/users", { email })
        .then((response) => {
          if (response.ok) {
            return resolve(response.json());
          } else if (response.status === 409) {
            throw new ConflictError();
          } else {
            throw new TechnicalError();
          }
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  /**
   * Fetches the current user.
   *
   * @return {Promise<User>}
   * @throws {UnauthorizedError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://teamhanko.github.io/hanko/#/authentication/IsUserAuthorized
   * @see https://teamhanko.github.io/hanko/#/management/listUser
   */
  getCurrent(): Promise<User> {
    return new Promise<User>((resolve, reject) =>
      this.client
        .get("/me")
        .then((response) => {
          if (response.ok) {
            return response.json();
          } else if (
            response.status === 400 ||
            response.status === 401 ||
            response.status === 404
          ) {
            throw new UnauthorizedError();
          } else {
            throw new TechnicalError();
          }
        })
        .then((me: Me) => {
          return this.client.get(`/users/${me.id}`);
        })
        .then((response) => {
          if (response.ok) {
            return resolve(response.json());
          } else if (
            response.status === 400 ||
            response.status === 401 ||
            response.status === 404
          ) {
            throw new UnauthorizedError();
          } else {
            throw new TechnicalError();
          }
        })
        .catch((e) => {
          reject(e);
        })
    );
  }
}

export { UserClient };
