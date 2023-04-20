import { SessionCreatedEventDetail, sessionRemovedType } from "./CustomEvents";
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
   * @param {SessionCreatedEventDetail} detail - The event detail.
   */
  private handleSessionCreatedEvent = (detail: SessionCreatedEventDetail) => {
    this._scheduler.removeTasksWithType(sessionRemovedType);
    this._scheduler.scheduleTask(
      sessionRemovedType,
      () => this.dispatchSessionRemovedEvent(),
      detail.expirationSeconds
    );
  };

  /**
   * Handles the "hanko-session-removed" event by removing scheduled "hanko-session-removed" events, to prevent it from
   * being triggered again (e.g. when there are multiple SDK instances).
   *
   * @private
   */
  private handleSessionRemovedEvent = () => {
    this._scheduler.removeTasksWithType(sessionRemovedType);
  };

  /**
   * Handles the "hanko-user-deleted" event by removing the scheduled "hanko-session-removed" events, because user
   * deletion implies that the session has also been removed.
   *
   * @private
   */
  private handleUserDeletedEvent = () => {
    this._scheduler.removeTasksWithType(sessionRemovedType);
  };

  /**
   * Returns the session detail currently stored in the local storage.
   *
   * @private
   * @returns {SessionCreatedEventDetail}
   */
  private getSessionDetail(): SessionCreatedEventDetail {
    this._sessionState.read();

    const userID = this._sessionState.getUserID();
    const expirationSeconds = this._sessionState.getExpirationSeconds();
    const jwt = this._cookie.getAuthCookie();

    return {
      userID,
      expirationSeconds,
      jwt,
    };
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

    const detail = this.getSessionDetail();

    if (detail.expirationSeconds <= 0) {
      this.dispatchSessionRemovedEvent();
      return;
    }

    this.dispatchSessionCreatedEvent(detail);
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
    this._listener.onSessionCreated(this.handleSessionCreatedEvent);
    this._listener.onSessionRemoved(this.handleSessionRemovedEvent);
    this._listener.onUserDeleted(this.handleUserDeletedEvent);

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

    const detail = this.getSessionDetail();

    if (detail.userID && detail.expirationSeconds > 0) {
      this.dispatchSessionCreatedEvent(detail);
    }
  }
}
