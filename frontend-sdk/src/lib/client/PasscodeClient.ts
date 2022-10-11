import { PasscodeState } from "../state/PasscodeState";
import { Passcode } from "../Dto";
import {
  InvalidPasscodeError,
  MaxNumOfPasscodeAttemptsReachedError,
  TechnicalError,
  TooManyRequestsError,
} from "../Errors";
import { Client } from "./Client";

/**
 * A class to handle passcodes.
 *
 * @constructor
 * @category SDK
 * @subcategory Clients
 * @extends {Client}
 */
class PasscodeClient extends Client {
  state: PasscodeState;

  // eslint-disable-next-line require-jsdoc
  constructor(api: string, timeout = 13000) {
    super(api, timeout);
    /**
     *  @public
     *  @type {PasscodeState}
     */
    this.state = new PasscodeState();
  }

  /**
   * Causes the API to send a new passcode to the user's email address.
   *
   * @param {string} userID - The UUID of the user.
   * @return {Promise<Passcode>}
   * @throws {TooManyRequestsError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api/public#tag/Passcode/operation/passcodeInit
   */
  async initialize(userID: string): Promise<Passcode> {
    const response = await this.client.post("/passcode/login/initialize", {
      user_id: userID,
    });

    if (response.status === 429) {
      const retryAfter = parseInt(
        response.headers.get("X-Retry-After") || "0",
        10
      );

      this.state.read().setResendAfter(userID, retryAfter).write();
      throw new TooManyRequestsError(retryAfter);
    } else if (!response.ok) {
      throw new TechnicalError();
    }

    const passcode = response.json();

    this.state
      .read()
      .setActiveID(userID, passcode.id)
      .setTTL(userID, passcode.ttl)
      .write();

    return passcode;
  }

  /**
   * Validates the passcode obtained from the email.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} code - The passcode digests.
   * @return {Promise<void>}
   * @throws {InvalidPasscodeError}
   * @throws {MaxNumOfPasscodeAttemptsReachedError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api/public#tag/Passcode/operation/passcodeFinal
   */
  async finalize(userID: string, code: string): Promise<void> {
    const passcodeID = this.state.read().getActiveID(userID);
    const response = await this.client.post("/passcode/login/finalize", {
      id: passcodeID,
      code,
    });

    if (response.status === 401) {
      throw new InvalidPasscodeError();
    } else if (response.status === 410) {
      this.state.reset(userID).write();
      throw new MaxNumOfPasscodeAttemptsReachedError();
    } else if (!response.ok) {
      throw new TechnicalError();
    }

    this.state.reset(userID).write();

    return;
  }

  /**
   * Returns the number of seconds the current passcode is active for.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getTTL(userID: string) {
    return this.state.read().getTTL(userID);
  }

  /**
   * Returns the number of seconds the rate limiting is active for.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getResendAfter(userID: string) {
    return this.state.read().getResendAfter(userID);
  }
}

export { PasscodeClient };
