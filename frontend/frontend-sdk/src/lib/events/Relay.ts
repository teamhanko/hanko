import { Listener } from "./Listener";
import { Dispatcher } from "./Dispatcher";
import { SessionClient } from "../client/SessionClient";
import { SessionState } from "./SessionState";
import { WindowActivityManager } from "./WindowActivityManager";
import { Scheduler, SessionCheckResult } from "./Scheduler";
import { SessionChannel, BroadcastMessage } from "./SessionChannel";
import { InternalOptions } from "../../Hanko";

/**
 * A class that manages session checks, dispatches events based on session status,
 * and uses broadcast channels for inter-tab communication.
 *
 * @category SDK
 * @subcategory Internal
 * @extends Dispatcher
 * @param {string} api - The API endpoint URL.
 * @param {InternalOptions} options - The internal configuration options of the SDK.
 */
export class Relay extends Dispatcher {
  listener = new Listener(); // Listener for session-related events.
  checkInterval = 30000; // Interval for session validity checks in milliseconds.
  client: SessionClient; // Client for session validation.
  sessionState: SessionState; // Manages session-related states.
  windowActivityManager: WindowActivityManager; // Manages window activity states.
  scheduler: Scheduler; //  Schedules session validity checks.
  sessionChannel: SessionChannel; // Handles inter-tab communication via broadcast channels.

  // eslint-disable-next-line require-jsdoc
  constructor(api: string, options: InternalOptions) {
    super();
    this.client = new SessionClient(api, options);
    this.checkInterval = options.sessionCheckInterval;
    this.sessionState = new SessionState(this.checkInterval);
    this.windowActivityManager = new WindowActivityManager(
      () => this.startSessionCheck(),
      () => this.stopSessionCheck(),
    );
    this.scheduler = new Scheduler(
      this.checkInterval,
      () => this.checkSession(),
      () => this.onSessionExpired(true),
    );
    this.sessionChannel = new SessionChannel(
      options.sessionCheckChannelName,
      () => this.onSessionExpired(true),
      (msg) => this.onSessionCreated(msg),
      () => this.onLeadershipRequested(),
      (msg) => this.onCheckCompleted(msg),
    );

    this.initializeEventListeners();
  }

  /**
   * Sets up all event listeners and initializes session management.
   * This method is crucial for ensuring the session is monitored across all tabs.
   * @private
   */
  private initializeEventListeners(): void {
    // Listen for session creation events
    this.listener.onSessionCreated((detail) => {
      if (this.sessionState.getIsLoggedIn()) return;
      this.sessionState.setIsLoggedIn(true);
      this.startSessionCheck(); // Begin session checks now that a user is logged in
      this.sessionChannel.post("sessionCreated", {
        claims: detail.claims,
      }); // Inform other tabs
    });

    // Listen for user logout events
    this.listener.onUserLoggedOut(() => {
      if (!this.sessionState.getIsLoggedIn()) return;
      this.sessionChannel.post("sessionExpired"); // Inform other tabs session ended
      this.onSessionExpired(false); // Reset session state
    });
  }

  /**
   * Initiates session checking based on the last check time.
   * This method decides when the next check should occur to balance between performance and freshness.
   * @private
   */
  private startSessionCheck(): void {
    this.sessionChannel.post("requestLeadership"); // Inform other tabs this tab is now checking
    if (this.scheduler.isRunning()) return;

    const now = Date.now();
    const expiresSoon = this.sessionState.isExpiringSoon(now);

    if (expiresSoon) {
      const timeToExpiration = this.sessionState.getTimeToExpiration(now);
      this.scheduler.sessionTimeoutAfter(timeToExpiration);
    } else {
      const timeToNextCheck = this.sessionState.getTimeToNextCheck(now);
      this.scheduler.start(timeToNextCheck);
    }
  }

  /**
   * Stops session checking.
   * @private
   */
  private stopSessionCheck(): void {
    this.scheduler.stop();
  }

  /**
   * Resets session-related states when a session becomes invalid.
   * This ensures all session-related variables are cleared to avoid stale data.
   * @private
   */
  private onSessionExpired(dispatchEvent: boolean = false) {
    if (!this.sessionState.getIsLoggedIn()) return;
    if (dispatchEvent) {
      this.dispatchSessionExpiredEvent(); // Inform listeners that session expired
    }
    this.scheduler.stop();
    this.sessionState.reset();
  }

  /**
   * Handles session creation events from broadcast messages.
   * @param {BroadcastMessage} msg - The broadcast message containing session details.
   * @private
   */
  private onSessionCreated(msg: BroadcastMessage) {
    if (this.sessionState.getIsLoggedIn()) return;
    const now = Date.now();
    const expiration = Date.parse(msg.claims.expiration);
    this.sessionState.setExpiration(expiration);
    const expirationSeconds = this.sessionState.getTimeToExpiration(now);
    this.sessionState.setIsLoggedIn(true);
    this.dispatchSessionCreatedEvent({
      claims: msg.claims,
      expirationSeconds, // deprecated
    }); // Notify listeners of new session
  }

  /**
   * Handles leadership requests from other tabs.
   * @private
   */
  private onLeadershipRequested() {
    if (this.windowActivityManager.hasFocus()) return; // Ignore leadership requests when the 'document' is still focused.
    this.stopSessionCheck();
  }

  /**
   * Handles completed session checks from broadcast messages.
   * @param {BroadcastMessage} msg - The broadcast message containing session check details.
   * @private
   */
  private onCheckCompleted(msg: BroadcastMessage) {
    // Update with latest session info
    this.sessionState.setExpiration(msg.sessionExpiration);
    this.sessionState.setLastCheck(msg.lastCheck);
  }

  /**
   * Validates the current session and updates session information.
   * This method checks if the session is still valid and updates local data accordingly.
   * @returns {Promise<{SessionCheckResult>} - A promise that resolves with the session check result.
   * @private
   */
  private async checkSession(): Promise<SessionCheckResult> {
    try {
      const now = Date.now();
      const sessionResponse = await this.client.validate();
      const sessionExpiration = Date.parse(sessionResponse.expiration_time);

      if (this.sessionState.getIsLoggedIn() && !sessionResponse.is_valid) {
        this.sessionChannel.post("sessionExpired"); // Inform other tabs
      }

      this.sessionState.setLastCheck(now);
      this.sessionState.setExpiration(sessionExpiration);
      this.sessionState.setIsLoggedIn(sessionResponse.is_valid);

      this.sessionChannel.post("checkCompleted", {
        sessionExpiration: this.sessionState.getExpiration(),
        lastCheck: this.sessionState.getLastCheck(),
      }); // Share latest session info

      const expiresSoon = this.sessionState.isExpiringSoon(now);
      const timeToExpiration = this.sessionState.getTimeToExpiration(now);

      return {
        expiresSoon,
        timeToExpiration,
        ...sessionResponse,
      };
    } catch (e) {
      console.error("Error during session validation:", e);
    }

    return null;
  }
}
