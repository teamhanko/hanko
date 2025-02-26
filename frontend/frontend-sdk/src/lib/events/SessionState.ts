/**
 * Represents the session state with expiration and last check timestamps.
 *
 * @category SDK
 * @subcategory Internal
 */
export interface State {
  expiration: number; // Timestamp (in milliseconds) when the session expires.
  lastCheck: number; // Timestamp (in milliseconds) of the last session check.
}

/**
 * Manages session state persistence using localStorage.
 *
 * @category SDK
 * @subcategory Internal
 */
export class SessionState {
  private readonly storageKey: string;
  private readonly defaultState: State = {
    expiration: 0,
    lastCheck: 0,
  };

  /**
   * Creates an instance of SessionState.
   *
   * @param {string} storageKey - The key used to store session state in localStorage.
   */
  constructor(storageKey: string) {
    this.storageKey = storageKey;
  }

  /**
   * Loads the current session state from localStorage.
   *
   * @returns {State} The parsed session state or a default state if not found.
   */
  load(): State {
    const item = window.localStorage.getItem(this.storageKey);
    return item == null ? this.defaultState : JSON.parse(item);
  }

  /**
   * Saves the session state to localStorage.
   *
   * @param {State | null} session - The session state to save. If null, the default state is used.
   */
  save(session: State | null): void {
    window.localStorage.setItem(
      this.storageKey,
      JSON.stringify(session ? session : this.defaultState),
    );
  }
}
