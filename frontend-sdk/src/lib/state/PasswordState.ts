import { State } from "./State";
import { UserState } from "./UserState";

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {number=} retryAfter - Timestamp until when the next password login can be attempted in seconds (since January 1, 1970 00:00:00 UTC).
 */
export interface LocalStoragePassword {
  retryAfter?: number;
}

/**
 * A class that manages the password login state.
 *
 * @extends UserState
 * @category SDK
 * @subcategory Internal
 */
class PasswordState extends UserState {
  /**
   * Get the password state.
   *
   * @private
   * @param {string} userID - The UUID of the user.
   * @return {LocalStoragePassword}
   */
  private getState(userID: string): LocalStoragePassword {
    return (super.getUserState(userID).password ||= {});
  }

  /**
   * Reads the current state.
   *
   * @public
   * @return {PasswordState}
   */
  read(): PasswordState {
    super.read();

    return this;
  }

  /**
   * Gets the number of seconds until when a new password login can be attempted.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getRetryAfter(userID: string): number {
    return State.timeToRemainingSeconds(this.getState(userID).retryAfter);
  }

  /**
   * Sets the number of seconds until a new password login can be attempted.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} seconds - Number of seconds the passcode is valid for.
   * @return {PasswordState}
   */
  setRetryAfter(userID: string, seconds: number): PasswordState {
    this.getState(userID).retryAfter = State.remainingSecondsToTime(seconds);

    return this;
  }
}

export { PasswordState };
