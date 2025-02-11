import { SessionCheckResponse } from "../Dto";

// Type representing data returned by the session check callback.
export type SessionCheckResult =
  | (SessionCheckResponse & { timeToExpiration: number; expiresSoon: boolean })
  | null;

// Callback type for performing a session check.
type SessionCheckCallback = () => Promise<SessionCheckResult>;

// Callback type for handling session timeout events.
type SessionExpiredCallback = () => void;

/**
 * Manages scheduling for periodic and timeout-based session checks.
 *
 * @category SDK
 * @subcategory Internal
 * @param {number} checkInterval - The interval in milliseconds between periodic session checks.
 * @param {SessionCheckCallback} onSessionCheck - The callback function to perform a session check.
 * @param {SessionExpiredCallback} onSessionExpired - The callback function to handle session timeout events.
 */
export class Scheduler {
  private intervalID: ReturnType<typeof setInterval> | null = null; // Identifier for the periodic check interval.
  private timeoutID: ReturnType<typeof setTimeout> | null = null; // Identifier for the session expiration timeout.
  private readonly checkInterval: number; // The interval between periodic session checks.
  private readonly onSessionCheck: SessionCheckCallback; // The callback function to perform a session check.
  private readonly onSessionExpired: SessionExpiredCallback; // The callback function to handle session expired events.

  // eslint-disable-next-line require-jsdoc
  constructor(
    checkInterval: number,
    onSessionCheck: SessionCheckCallback,
    onSessionExpired: SessionExpiredCallback,
  ) {
    this.checkInterval = checkInterval;
    this.onSessionCheck = onSessionCheck;
    this.onSessionExpired = onSessionExpired;

    this.start(this.checkInterval);
  }

  /**
   * Handles the session expiration when it is about to expire soon.
   * Stops any ongoing checks and schedules a timeout for the expiration.
   *
   * @param {number} timeToExpiration - The time in milliseconds until the session expires.
   */
  sessionTimeoutAfter(timeToExpiration: number): void {
    this.stop();
    this.timeoutID = setTimeout(async () => {
      this.onSessionExpired();
    }, timeToExpiration);
  }

  /**
   * Starts the session check process.
   * Schedules the first check after an optional initial delay and begins periodic checks.
   *
   * @param {number} initialDelay - The delay in milliseconds before the first check is performed.
   */
  start(initialDelay: number = 0): void {
    // Schedule the first check after an optional delay
    this.timeoutID = setTimeout(async () => {
      let sessionCheckResult = await this.onSessionCheck();

      if (sessionCheckResult.is_valid) {
        if (sessionCheckResult.expiresSoon) {
          this.sessionTimeoutAfter(sessionCheckResult.timeToExpiration);
          return;
        }

        // Begin periodic checks
        this.intervalID = setInterval(async () => {
          sessionCheckResult = await this.onSessionCheck();

          if (sessionCheckResult.is_valid) {
            if (sessionCheckResult.expiresSoon) {
              this.sessionTimeoutAfter(sessionCheckResult.timeToExpiration);
            }
          } else {
            this.stop();
            this.onSessionExpired();
          }
        }, this.checkInterval);
      } else {
        this.stop();
        this.onSessionExpired();
      }
    }, initialDelay);
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
}
