import { State } from "../State";

/**
 * Options for SessionState
 *
 * @category SDK
 * @subcategory Internal
 * @property {string} localStorageKey - The prefix / name of the local storage keys.
 */
interface SessionStateOptions {
  localStorageKey: string;
}

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {Object.<string, LocalStorageSession>} - A dictionary for mapping users to their states.
 */
export interface LocalStorageSession {
  expiry: number;
  authFlowCompleted: boolean;
}

/**
 * A class to read and write local storage contents regarding sessions.
 *
 * @extends State
 * @param {SessionStateOptions} options - The options that can be used
 * @category SDK
 * @subcategory Internal
 */
class SessionState extends State {
  // eslint-disable-next-line require-jsdoc
  constructor(options: SessionStateOptions) {
    super(`${options.localStorageKey}_session`);
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
    this.ls.session ||= { expiry: 0, authFlowCompleted: false };
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
   * Gets the authFlowCompleted indicator.
   */
  getAuthFlowCompleted(): boolean {
    return this.getState().authFlowCompleted;
  }

  /**
   * Sets the authFlowCompleted indicator.
   *
   * @param {string} completed - The authFlowCompleted indicator.
   * @return {SessionState}
   */
  setAuthFlowCompleted(completed: boolean): SessionState {
    this.getState().authFlowCompleted = completed;
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

    return this;
  }
}

export { SessionState };
