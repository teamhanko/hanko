import { PasswordState } from "../state/PasswordState";
import { PasscodeState } from "../state/PasscodeState";
import {
  InvalidPasswordError,
  TechnicalError,
  TooManyRequestsError,
  UnauthorizedError,
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
  passwordState: PasswordState;
  passcodeState: PasscodeState;

  // eslint-disable-next-line require-jsdoc
  constructor(api: string, timeout = 13000) {
    super(api, timeout);
    /**
     *  @public
     *  @type {PasswordState}
     */
    this.passwordState = new PasswordState();
    /**
     *  @public
     *  @type {PasscodeState}
     */
    this.passcodeState = new PasscodeState();
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
   * @see https://docs.hanko.io/api/public#tag/Password/operation/passwordLogin
   */
  async login(userID: string, password: string): Promise<void> {
    const response = await this.client.post("/password/login", {
      user_id: userID,
      password,
    });

    if (response.status === 401) {
      throw new InvalidPasswordError();
    } else if (response.status === 429) {
      const retryAfter = response.parseXRetryAfterHeader();
      this.passwordState.read().setRetryAfter(userID, retryAfter).write();
      throw new TooManyRequestsError(retryAfter);
    } else if (!response.ok) {
      throw new TechnicalError();
    }

    this.passcodeState.read().reset(userID).write();

    return;
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
   * @see https://docs.hanko.io/api/public#tag/Password/operation/password
   */
  async update(userID: string, password: string): Promise<void> {
    const response = await this.client.put("/password", {
      user_id: userID,
      password,
    });

    if (response.status === 401) {
      throw new UnauthorizedError();
    } else if (!response.ok) {
      throw new TechnicalError();
    }

    return;
  }

  /**
   * Returns the number of seconds the rate limiting is active for.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getRetryAfter(userID: string) {
    return this.passwordState.read().getRetryAfter(userID);
  }
}

export { PasswordClient };
