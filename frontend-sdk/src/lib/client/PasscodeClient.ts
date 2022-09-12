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
  private state: PasscodeState;

  // eslint-disable-next-line require-jsdoc
  constructor(api: string, timeout: number) {
    super(api, timeout);
    /**
     *  @private
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
   * @see https://docs.hanko.io/api#tag/Passcode/operation/passcodeInit
   */
  initialize(userID: string): Promise<Passcode> {
    return new Promise<Passcode>((resolve, reject) => {
      this.client
        .post("/passcode/login/initialize", { user_id: userID })
        .then((response) => {
          if (response.ok) {
            return response.json();
          } else if (response.status === 429) {
            const retryAfter = parseInt(
              response.headers.get("X-Retry-After") || "0",
              10
            );

            this.state.read().setResendAfter(userID, retryAfter).write();

            throw new TooManyRequestsError(retryAfter);
          } else {
            throw new TechnicalError();
          }
        })
        .then((passcode: Passcode) => {
          this.state
            .read()
            .setActiveID(userID, passcode.id)
            .setTTL(userID, passcode.ttl)
            .write();
          return resolve(passcode);
        })
        .catch((e) => {
          reject(e);
        });
    });
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
   * @see https://docs.hanko.io/api#tag/Passcode/operation/passcodeFinal
   */
  finalize(userID: string, code: string): Promise<void> {
    const passcodeID = this.state.read().getActiveID(userID);

    return new Promise<void>((resolve, reject) => {
      this.client
        .post("/passcode/login/finalize", { id: passcodeID, code })
        .then((response) => {
          if (response.ok) {
            this.state.reset(userID).write();

            return resolve();
          } else if (response.status === 401) {
            throw new InvalidPasscodeError();
          } else if (response.status === 404 || response.status === 410) {
            this.state.reset(userID).write();

            throw new MaxNumOfPasscodeAttemptsReachedError();
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
