import { SessionCheckResponse } from "../Dto";

// Type representing data returned by the session check callback.
export type SessionCheckResult =
  | (Omit<SessionCheckResponse, "expiration_time"> & {
      expiration: number;
    })
  | null;

/**
 * Callback type for performing a session check.
 * @ignore
 */
type SessionCheckCallback = () => Promise<SessionCheckResult>;

/**
 * Callback type for handling session timeout events.
 * @ignore
 */
type SessionExpiredCallback = () => void;

/**
 * Manages scheduling for periodic and timeout-based session checks.
 *
 * @category SDK
 * @subcategory Internal
 * @param {number} checkInterval - The interval in milliseconds between periodic session checks.
 * @param {SessionCheckCallback} checkSession - The callback function to perform a session check.
 * @param {SessionExpiredCallback} onSessionExpired - The callback function to handle session timeout events.
 */
export class Scheduler {
  private intervalID: ReturnType<typeof setInterval> | null = null; // Identifier for the periodic check interval.
  private timeoutID: ReturnType<typeof setTimeout> | null = null; // Identifier for the session expiration timeout.
  private readonly checkInterval: number; // The interval between periodic session checks.
  private readonly checkSession: SessionCheckCallback; // The callback function to perform a session check.
  private readonly onSessionExpired: SessionExpiredCallback; // The callback function to handle session expired events.

  // eslint-disable-next-line require-jsdoc
  constructor(
    checkInterval: number,
    checkSession: SessionCheckCallback,
    onSessionExpired: SessionExpiredCallback,
  ) {
    this.checkInterval = checkInterval;
    this.checkSession = checkSession;
    this.onSessionExpired = onSessionExpired;
  }

  /**
   * Handles the session expiration when it is about to expire soon.
   * Stops any ongoing checks and schedules a timeout for the expiration.
   *
   * @param {number} timeToExpiration - The time in milliseconds until the session expires.
   */
  scheduleSessionExpiry(timeToExpiration: number): void {
    this.stop();
    this.timeoutID = setTimeout(async () => {
      this.stop();
      this.onSessionExpired();
    }, timeToExpiration);
  }

  /**
   * Starts the session check process.
   * Determines when the next check should run based on the last known check time and session expiration.
   * If the session is expiring soon, schedules an expiration event instead of starting periodic checks.
   *
   * @param {number} lastCheck - The timestamp (in milliseconds) of the last session check.
   * @param {number} expiration - The timestamp (in milliseconds) of when the session expires.
   */
  start(lastCheck: number = 0, expiration: number = 0): void {
    const timeToNextCheck = this.calcTimeToNextCheck(lastCheck);

    if (this.sessionExpiresSoon(expiration)) {
      this.scheduleSessionExpiry(timeToNextCheck);
      return;
    }

    // Schedule the first check after an optional delay
    this.timeoutID = setTimeout(async () => {
      let result = await this.checkSession();

      if (result.is_valid) {
        if (this.sessionExpiresSoon(result.expiration)) {
          this.scheduleSessionExpiry(result.expiration - Date.now());
          return;
        }

        // Begin periodic checks
        this.intervalID = setInterval(async () => {
          result = await this.checkSession();

          if (result.is_valid) {
            if (this.sessionExpiresSoon(result.expiration)) {
              this.scheduleSessionExpiry(result.expiration - Date.now());
            }
          } else {
            this.stop();
          }
        }, this.checkInterval);
      } else {
        this.stop();
      }
    }, timeToNextCheck);
  }

  /**
   * Stops the session check process and clears all timers.
   */
  stop(): void {
    if (this.timeoutID) {
      clearTimeout(this.timeoutID);
      this.timeoutID = null;
    }

    if (this.intervalID) {
      clearInterval(this.intervalID);
      this.intervalID = null;
    }
  }

  /**
   * Checks if the scheduler is currently running.
   * @returns {boolean} True if the scheduler is running; otherwise, false.
   */
  isRunning(): boolean {
    return this.timeoutID !== null || this.intervalID !== null;
  }
  /**
   * Checks if the session is about to expire.
   * @param {number} expiration - Timestamp when the session will expire.
   * @returns {boolean} True if the session is about to expire; otherwise, false.
   */
  sessionExpiresSoon(expiration: number): boolean {
    return expiration > 0 && expiration - Date.now() <= this.checkInterval;
  }

  /**
   * Calculates the time until the next session check should occur.
   *
   * @param {number} lastCheck - The timestamp (in milliseconds) of the last session check.
   * @returns {number} The time in milliseconds until the next check should be performed.
   */
  calcTimeToNextCheck(lastCheck: number): number {
    const timeSinceLastCheck = Date.now() - lastCheck;
    return this.checkInterval >= timeSinceLastCheck
      ? this.checkInterval - (timeSinceLastCheck % this.checkInterval)
      : 0;
  }
}
