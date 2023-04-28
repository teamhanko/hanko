import { State } from "../State";

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {Object.<string, LocalStorageSession>} - A dictionary for mapping users to their states.
 */
export interface LocalStorageSession {
  expiry: number;
  userID: string;
}

/**
 * A class to read and write local storage contents regarding sessions.
 *
 * @extends State
 * @param {string} key - The local storage key.
 * @category SDK
 * @subcategory Internal
 */
class SessionState extends State {
  // eslint-disable-next-line require-jsdoc
  constructor() {
    super("hanko_session");
  }

  /**
   * Reads the current state.
   *
   * @public
   * @return {SessionState}
   */
  read(): SessionState {
    super.read();

    return this;
  }

  /**
   * Gets the session state.
   *
   * @return {LocalStorageSession}
   */
  getState(): LocalStorageSession {
    this.ls.session ||= { expiry: 0, userID: "" };
    return this.ls.session;
  }

  /**
   * Gets the number of seconds until the active session is valid.
   *
   * @return {number}
   */
  getExpirationSeconds(): number {
    return State.timeToRemainingSeconds(this.getState().expiry);
  }

  /**
   * Sets the number of seconds until the active session is valid.
   *
   * @param {number} seconds - The number of seconds
   * @return {SessionState}
   */
  setExpirationSeconds(seconds: number): SessionState {
    this.getState().expiry = State.remainingSecondsToTime(seconds);
    return this;
  }

  /**
   * Gets the user id.
   */
  getUserID(): string {
    return this.getState().userID;
  }

  /**
   * Sets the user id.
   *
   * @param {string} userID - The user id
   * @return {SessionState}
   */
  setUserID(userID: string): SessionState {
    this.getState().userID = userID;
    return this;
  }

  /**
   * Removes the session details.
   *
   * @return {SessionState}
   */
  reset(): SessionState {
    const session = this.getState();

    delete session.expiry;
    delete session.userID;

    return this;
  }
}

export { SessionState };
