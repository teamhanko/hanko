import { PasswordState } from "../state/PasswordState";
import {
  InvalidPasswordError,
  TechnicalError,
  TooManyRequestsError,
} from "../Errors";
import { Client } from "./Client";

/**
 * A class to handle passwords.
 *
 * @constructor
 * @category SDK
 * @subcategory Clients
 * @extends {Client}
 */
class PasswordClient extends Client {
  private state: PasswordState;

  // eslint-disable-next-line require-jsdoc
  constructor(api: string, timeout: number) {
    super(api, timeout);
    /**
     *  @private
     *  @type {PasswordState}
     */
    this.state = new PasswordState();
  }

  /**
   * Logs in a user with a password.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} password - The password.
   * @return {Promise<void>}
   * @throws {TooManyRequestsError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api#tag/Password/operation/passwordLogin
   */
  login(userID: string, password: string): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      this.client
        .post("/password/login", { user_id: userID, password })
        .then((response) => {
          if (response.ok) {
            return resolve();
          } else if (response.status === 401) {
            throw new InvalidPasswordError();
          } else if (response.status === 429) {
            const retryAfter = parseInt(
              response.headers.get("X-Retry-After") || "0",
              10
            );

            this.state.read().setRetryAfter(userID, retryAfter).write();

            throw new TooManyRequestsError(retryAfter);
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
   * Updates a password.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} password - The new password.
   * @return {Promise<void>}
   * @throws {RequestTimeoutError}
   * @throws {UnauthorizedError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api#tag/Password/operation/password
   */
  update(userID: string, password: string): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      this.client
        .put("/password", { user_id: userID, password })
        .then((response) => {
          if (response.ok) {
            return resolve();
          }

          throw new TechnicalError();
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  /**
   * Returns the number of seconds the rate limiting is active for.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getRetryAfter(userID: string) {
    return this.state.read().getRetryAfter(userID);
  }
}

export { PasswordClient };
