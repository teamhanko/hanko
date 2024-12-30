/**
 * Manages session-related operations such as validation, expiration, and reset.
 *
 * @category SDK
 * @subcategory Internal
 */
export class SessionState {
  private lastCheck: number = Date.now(); // Timestamp of the last session validity check.
  private expiration: number = 0; // Timestamp when the current session will expire.
  private isLoggedIn: boolean = false; // Indicates if a user is currently logged in.
  private readonly checkInterval: number; // milliseconds; interval for session validity checks.

  // eslint-disable-next-line require-jsdoc
  constructor(checkInterval: number) {
    this.checkInterval = checkInterval;
  }

  /**
   * Checks if the user is currently logged in.
   * @returns {boolean} True if the user is logged in; otherwise, false.
   */
  getIsLoggedIn(): boolean {
    return this.isLoggedIn;
  }

  /**
   * Sets the login status of the user.
   * @param {boolean} loggedIn - True if the user is logged in; otherwise, false.
   */
  setIsLoggedIn(loggedIn: boolean): void {
    this.isLoggedIn = loggedIn;
  }

  /**
   * Retrieves the last session check timestamp.
   * @returns {number} The timestamp of the last session check.
   */
  getLastCheck(): number {
    return this.lastCheck;
  }

  /**
   * Updates the timestamp of the last session check.
   * @param {number} timestamp - The new timestamp.
   */
  setLastCheck(timestamp: number): void {
    this.lastCheck = timestamp;
  }

  /**
   * Retrieves the session expiration time.
   * @returns {number} The session expiration timestamp.
   */
  getExpiration(): number {
    return this.expiration;
  }

  /**
   * Sets the session expiration time.
   * @param {number} expiration - The new expiration timestamp.
   */
  setExpiration(expiration: number): void {
    this.expiration = expiration;
  }

  /**
   * Resets the session state.
   */
  reset(): void {
    this.isLoggedIn = false;
    this.expiration = 0;
  }

  /**
   * Checks if the session is about to expire.
   * @param {number} now - The current timestamp.
   * @returns {boolean} True if the session is about to expire; otherwise, false.
   */
  isExpiringSoon(now: number): boolean {
    return this.expiration > 0 && this.expiration - now <= this.checkInterval;
  }

  /**
   * Retrieves the time remaining until the session expires.
   * @param {number} now - The current timestamp.
   * @returns {number} The time remaining until session expiration in milliseconds.
   */
  getTimeToExpiration(now: number): number {
    return this.expiration - now;
  }

  /**
   * Retrieves the time elapsed since the last session check.
   * @param {number} now - The current timestamp.
   * @returns {number} The time elapsed since the last check in milliseconds.
   */
  getTimeSinceLastCheck(now: number): number {
    return now - this.lastCheck;
  }

  /**
   * Calculates the time until the next session check should occur.
   * @param {number} now - The current timestamp.
   * @returns {number} The time remaining until the next check in milliseconds.
   */
  getTimeToNextCheck(now: number): number {
    const timeSinceLastCheck = this.getTimeSinceLastCheck(now);
    return this.checkInterval - (timeSinceLastCheck % this.checkInterval);
  }
}
