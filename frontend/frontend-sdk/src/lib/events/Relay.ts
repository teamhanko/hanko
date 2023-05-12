import { SessionEventDetail, sessionExpiredType } from "./CustomEvents";
import { Listener } from "./Listener";
import { SessionState } from "../state/session/SessionState";
import { Scheduler } from "./Scheduler";
import { Dispatcher } from "./Dispatcher";
import { Cookie } from "../Cookie";

/**
 * A class that dispatches events and scheduled events, based on other events.
 *
 * @category SDK
 * @subcategory Internal
 * @extends Dispatcher
 */
export class Relay extends Dispatcher {
  _listener = new Listener();
  _scheduler = new Scheduler();
  _sessionState = new SessionState();
  _cookie = new Cookie();

  // eslint-disable-next-line require-jsdoc
  constructor() {
    super();
    this.listenEventDependencies();
  }

  /**
   * Removes the scheduled "hanko-session-removed" event and re-schedules a new event with updated expirationSeconds, to
   * ensure the "hanko-session-removed" event won't be triggered too early.
   *
   * @private
   * @param {SessionEventDetail} detail - The event detail.
   */
  private scheduleSessionExpiredEvent = (detail: SessionEventDetail) => {
    this._scheduler.removeTasksWithType(sessionExpiredType);
    this._scheduler.scheduleTask(
      sessionExpiredType,
      () => this.dispatchSessionExpiredEvent(),
      detail.expirationSeconds
    );
  };

  /**
   * Cancels scheduled "hanko-session-expired" events, to prevent it from being triggered again (e.g. when there are
   * multiple SDK instances).
   *
   * @private
   */
  private cancelSessionExpiredEvent = () => {
    this._scheduler.removeTasksWithType(sessionExpiredType);
  };

  /**
   * Returns the session detail currently stored in the local storage.
   *
   * @returns {SessionEventDetail}
   */
  public getSessionDetail(): SessionEventDetail {
    this._sessionState.read();

    const userID = this._sessionState.getUserID();
    const expirationSeconds = this._sessionState.getExpirationSeconds();
    const jwt = this._cookie.getAuthCookie();

    return expirationSeconds > 0 && userID.length
      ? {
          userID,
          expirationSeconds,
          jwt,
        }
      : null;
  }

  /**
   * Handles the "storage" event in case the local storage entry, that contains the session detail has been changed by
   * another window. Depending on the new value of `expirationSeconds`, it either dispatches a "hanko-session-created"
   * or a "hanko-session-removed" event.
   *
   * @private
   * @param {StorageEvent} event - The storage event object.
   */
  private handleStorageEvent = (event: StorageEvent) => {
    if (event.key !== "hanko_session") return;

    const sessionDetail = this.getSessionDetail();

    if (!sessionDetail) {
      this.dispatchSessionExpiredEvent();
      return;
    }

    this.dispatchSessionCreatedEvent(sessionDetail);
  };

  /**
   * Calls `func` when the document is ready.
   *
   * @param {function} func - The function to be called.
   * @private
   */
  private static onDocumentReady(func: () => any) {
    if (
      document.readyState === "complete" ||
      document.readyState === "interactive"
    ) {
      setTimeout(func, 1);
    } else {
      document.addEventListener("DOMContentLoaded", func);
    }
  }

  /**
   * Listens for events sent in the current browser.
   *
   * @private
   */
  private listenEventDependencies() {
    this._listener.onSessionCreated(this.scheduleSessionExpiredEvent);
    this._listener.onSessionResumed(this.scheduleSessionExpiredEvent);
    this._listener.onSessionExpired(this.cancelSessionExpiredEvent);
    this._listener.onUserDeleted(this.cancelSessionExpiredEvent);

    // Handle cases, where the session has been changed by another window.
    window.addEventListener("storage", this.handleStorageEvent);

    // Dispatch initial events once the document has been loaded, so all event binding should already be installed and
    // it`s not too early to call the `dispatchInitialEvents()` function.
    Relay.onDocumentReady(() => this.dispatchInitialEvents());
  }

  /**
   * Restores the previous session details and dispatches initial events.
   */
  public dispatchInitialEvents() {
    this._sessionState.read();

    const sessionDetail = this.getSessionDetail();

    if (sessionDetail) {
      this.dispatchSessionResumedEvent(sessionDetail);
    } else {
      this.dispatchSessionNotPresent();
    }
  }
}
