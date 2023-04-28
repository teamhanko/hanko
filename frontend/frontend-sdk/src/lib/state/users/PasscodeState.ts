import { State } from "../State";
import { UserState } from "./UserState";

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {string=} id - The UUID of the active passcode.
 * @property {number=} ttl - Timestamp until when the passcode is valid in seconds (since January 1, 1970 00:00:00 UTC).
 * @property {number=} resendAfter - Seconds until a passcode can be resent.
 * @property {emailID=} emailID - The email address ID.
 */
export interface LocalStoragePasscode {
  id?: string;
  ttl?: number;
  resendAfter?: number;
  emailID?: string;
}

/**
 * A class that manages passcodes via local storage.
 *
 * @extends UserState
 * @category SDK
 * @subcategory Internal
 */
class PasscodeState extends UserState {
  /**
   * Get the passcode state.
   *
   * @private
   * @param {string} userID - The UUID of the user.
   * @return {LocalStoragePasscode}
   */
  private getState(userID: string): LocalStoragePasscode {
    return (super.getUserState(userID).passcode ||= {});
  }

  /**
   * Reads the current state.
   *
   * @public
   * @return {PasscodeState}
   */
  read(): PasscodeState {
    super.read();

    return this;
  }

  /**
   * Gets the UUID of the active passcode.
   *
   * @param {string} userID - The UUID of the user.
   * @return {string}
   */
  getActiveID(userID: string): string {
    return this.getState(userID).id;
  }

  /**
   * Sets the UUID of the active passcode.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} passcodeID - The UUID of the passcode to be set as active.
   * @return {PasscodeState}
   */
  setActiveID(userID: string, passcodeID: string): PasscodeState {
    this.getState(userID).id = passcodeID;

    return this;
  }

  /**
   * Gets the UUID of the email address.
   *
   * @param {string} userID - The UUID of the user.
   * @return {string}
   */
  getEmailID(userID: string): string {
    return this.getState(userID).emailID;
  }

  /**
   * Sets the UUID of the email address.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} emailID - The UUID of the email address.
   * @return {PasscodeState}
   */
  setEmailID(userID: string, emailID: string): PasscodeState {
    this.getState(userID).emailID = emailID;

    return this;
  }

  /**
   * Removes the active passcode.
   *
   * @param {string} userID - The UUID of the user.
   * @return {PasscodeState}
   */
  reset(userID: string): PasscodeState {
    const passcode = this.getState(userID);

    delete passcode.id;
    delete passcode.ttl;
    delete passcode.resendAfter;
    delete passcode.emailID;

    return this;
  }

  /**
   * Gets the TTL in seconds. When the seconds expire, the code is invalid.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getTTL(userID: string): number {
    return State.timeToRemainingSeconds(this.getState(userID).ttl);
  }

  /**
   * Sets the passcode's TTL and stores it to the local storage.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} seconds - Number of seconds the passcode is valid for.
   * @return {PasscodeState}
   */
  setTTL(userID: string, seconds: number): PasscodeState {
    this.getState(userID).ttl = State.remainingSecondsToTime(seconds);

    return this;
  }

  /**
   * Gets the number of seconds until when the next passcode can be sent.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getResendAfter(userID: string): number {
    return State.timeToRemainingSeconds(this.getState(userID).resendAfter);
  }

  /**
   * Sets the number of seconds until a new passcode can be sent.
   *
   * @param {string} userID - The UUID of the user.
   * @param {number} seconds - Number of seconds the passcode is valid for.
   * @return {PasscodeState}
   */
  setResendAfter(userID: string, seconds: number): PasscodeState {
    this.getState(userID).resendAfter = State.remainingSecondsToTime(seconds);

    return this;
  }
}

export { PasscodeState };
