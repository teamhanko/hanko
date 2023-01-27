import { PasscodeState } from "../state/PasscodeState";
import { Passcode } from "../Dto";
import {
  InvalidPasscodeError,
  MaxNumOfPasscodeAttemptsReachedError,
  PasscodeExpiredError,
  TechnicalError,
  TooManyRequestsError,
  UnauthorizedError,
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
   * @param {string=} emailID - The UUID of the email address. If unspecified, the email will be sent to the primary email address.
   * @param {boolean=} force - Indicates the passcode should be sent, even if there is another active passcode.
   * @return {Promise<Passcode>}
   * @throws {TooManyRequestsError}
   * @throws {RequestTimeoutError}
   * @throws {UnauthorizedError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api/public#tag/Passcode/operation/passcodeInit
   */
  async initialize(
    userID: string,
    emailID?: string,
    force?: boolean
  ): Promise<Passcode> {
    this.state.read();

    const lastPasscodeTTL = this.state.getTTL(userID);
    const lastPasscodeID = this.state.getActiveID(userID);
    const lastEmailID = this.state.getEmailID(userID);
    let retryAfter = this.state.getResendAfter(userID);

    if (!force && lastPasscodeTTL > 0 && emailID === lastEmailID) {
      return {
        id: lastPasscodeID,
        ttl: lastPasscodeTTL,
      };
    }

    if (retryAfter > 0) {
      throw new TooManyRequestsError(retryAfter);
    }

    const body: any = { user_id: userID };

    if (emailID) {
      body.email_id = emailID;
    }

    const response = await this.client.post(`/passcode/login/initialize`, body);

    if (response.status === 429) {
      retryAfter = response.parseRetryAfterHeader();
      this.state.setResendAfter(userID, retryAfter).write();
      throw new TooManyRequestsError(retryAfter);
    } else if (response.status === 401) {
      throw new UnauthorizedError();
    } else if (!response.ok) {
      throw new TechnicalError();
    }

    const passcode: Passcode = response.json();

    this.state.setActiveID(userID, passcode.id).setTTL(userID, passcode.ttl);

    if (emailID) {
      this.state.setEmailID(userID, emailID);
    }

    this.state.write();

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
    const ttl = this.state.getTTL(userID);

    if (ttl <= 0) {
      throw new PasscodeExpiredError();
    }

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
