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
  private readonly checkInterval: number = 30000; // Interval for session validity checks in milliseconds.
  private readonly client: SessionClient; // Client for session validation.
  private readonly sessionState: SessionState; // Manages session-related states.
  private readonly windowActivityManager: WindowActivityManager; // Manages window activity states.
  private readonly scheduler: Scheduler; //  Schedules session validity checks.
  private readonly sessionChannel: SessionChannel; // Handles inter-tab communication via broadcast channels.
  private isLoggedIn: boolean;

  // eslint-disable-next-line require-jsdoc
  constructor(api: string, options: InternalOptions) {
    super();
    this.client = new SessionClient(api, options);
    this.checkInterval = options.sessionCheckInterval;
    this.sessionState = new SessionState(`${options.cookieName}_session_state`);
    this.sessionChannel = new SessionChannel(
      options.sessionCheckChannelName,
      () => this.onChannelSessionExpired(),
      (msg) => this.onChannelSessionCreated(msg),
      () => this.onChannelLeadershipRequested(),
    );
    this.scheduler = new Scheduler(
      this.checkInterval,
      () => this.checkSession(),
      () => this.onSessionExpired(),
    );
    this.windowActivityManager = new WindowActivityManager(
      () => this.startSessionCheck(),
      () => this.scheduler.stop(),
    );

    const now = Date.now();
    const { expiration } = this.sessionState.load();

    this.isLoggedIn = now < expiration;
    this.initializeEventListeners();
    this.startSessionCheck();
  }

  /**
   * Sets up all event listeners and initializes session management.
   * This method is crucial for ensuring the session is monitored across all tabs.
   * @private
   */
  private initializeEventListeners(): void {
    // Listen for session creation events
    this.listener.onSessionCreated((detail) => {
      const { claims } = detail;
      const expiration = Date.parse(claims.expiration);
      const lastCheck = Date.now();

      this.sessionState.save({ expiration, lastCheck }); // Save initial session state
      this.sessionChannel.post("sessionCreated", { claims }); // Inform other tabs
      this.startSessionCheck(); // Begin session checks now that a user is logged in
    });

    // Listen for user logout events
    this.listener.onUserLoggedOut(() => {
      this.sessionChannel.post("sessionExpired"); // Inform other tabs session ended
      this.sessionState.save(null);
      this.scheduler.stop();
    });

    window.addEventListener("beforeunload", () => this.scheduler.stop());
  }

  /**
   * Initiates session checking based on the last check time.
   * This method decides when the next check should occur to balance between performance and freshness.
   * @private
   */
  private startSessionCheck(): void {
    if (this.windowActivityManager.hasFocus()) {
      this.sessionChannel.post("requestLeadership"); // Inform other tabs this tab is now checking
    } else {
      return;
    }

    if (this.scheduler.isRunning()) {
      return;
    }

    const { lastCheck, expiration } = this.sessionState.load();

    if (this.isLoggedIn) {
      this.scheduler.start(lastCheck, expiration);
    }
  }

  /**
   * Validates the current session and updates session information.
   * This method checks if the session is still valid and updates local data accordingly.
   * @returns {Promise<{SessionCheckResult>} - A promise that resolves with the session check result.
   * @private
   */
  private async checkSession(): Promise<SessionCheckResult> {
    const lastCheck = Date.now();
    // eslint-disable-next-line camelcase
    const { is_valid, claims, expiration_time } = await this.client.validate();

    // eslint-disable-next-line camelcase
    const expiration = expiration_time ? Date.parse(expiration_time) : 0;

    // eslint-disable-next-line camelcase
    if (!is_valid && this.isLoggedIn) {
      this.dispatchSessionExpiredEvent();
    }

    // eslint-disable-next-line camelcase
    if (is_valid) {
      this.isLoggedIn = true;
      this.sessionState.save({ lastCheck, expiration });
    } else {
      this.isLoggedIn = false;
      this.sessionState.save(null);
      this.sessionChannel.post("sessionExpired"); // Inform other tabs
    }

    return {
      // eslint-disable-next-line camelcase
      is_valid,
      claims,
      expiration,
    };
  }

  /**
   * Resets session-related states when a session expires.
   * Ensures that authentication state is cleared and an expiration event is dispatched.
   * Assumes the user is logged out by default if the session state is unknown.
   * @private
   */
  private onSessionExpired() {
    if (this.isLoggedIn) {
      this.isLoggedIn = false;
      this.sessionState.save(null);
      this.sessionChannel.post("sessionExpired"); // Inform other tabs
      this.dispatchSessionExpiredEvent();
    }
  }

  /**
   * Handles session expired events from broadcast messages.
   * @private
   */
  private onChannelSessionExpired() {
    if (this.isLoggedIn) {
      this.isLoggedIn = false;
      this.dispatchSessionExpiredEvent();
    }
  }

  /**
   * Handles session creation events from broadcast messages.
   * @param {BroadcastMessage} msg - The broadcast message containing session details.
   * @private
   */
  private onChannelSessionCreated(msg: BroadcastMessage) {
    const { claims } = msg;
    const now = Date.now();
    const expiration = Date.parse(claims.expiration);
    const expirationSeconds = expiration - now;

    this.isLoggedIn = true;
    this.dispatchSessionCreatedEvent({
      claims,
      expirationSeconds, // deprecated
    });
  }

  /**
   * Handles leadership requests from other tabs.
   * @private
   */
  private onChannelLeadershipRequested() {
    if (!this.windowActivityManager.hasFocus()) {
      this.scheduler.stop();
    }
  }
}
