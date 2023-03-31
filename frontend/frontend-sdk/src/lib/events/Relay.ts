import { SessionCreatedEventDetail, sessionRemovedType } from "./CustomEvents";
import { Listener } from "./Listener";
import { SessionState } from "../state/session/SessionState";
import { Scheduler } from "./Scheduler";
import { Dispatcher } from "./Dispatcher";

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

  // eslint-disable-next-line require-jsdoc
  constructor() {
    super();
    this.listenEventDependencies();
    this.dispatchInitialEvents();
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
   * Handles the "storage" event in case the local storage entry, that contains the session detail has been changed by
   * another window. Depending on the new value of `expirationSeconds`, it either dispatches a "hanko-session-created"
   * or a "hanko-session-removed" event.
   *
   * @private
   * @param {StorageEvent} event - The storage event object.
   */
  private handleStorageEvent = (event: StorageEvent) => {
    if (event.key !== "hanko_session") return;

    const detail = this.getSessionDetailFromLocalStorage();

    if (detail.expirationSeconds <= 0) {
      this.dispatchSessionRemovedEvent();
      return;
    }

    this.dispatchSessionCreatedEvent(detail);
  };

  /**
   * Returns the session detail currently stored in the local storage.
   *
   * @private
   * @returns {SessionCreatedEventDetail}
   */
  private getSessionDetailFromLocalStorage(): SessionCreatedEventDetail {
    this._sessionState.read();

    const expirationSeconds = this._sessionState.getExpirationSeconds();
    const jwt = this._sessionState.getJWT();
    const userID = this._sessionState.getUserID();

    return {
      userID,
      expirationSeconds,
      jwt,
    };
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
  }

  /**
   * Restores the previous session details and dispatches initial events.
   */
  public dispatchInitialEvents() {
    this._sessionState.read();

    const detail = this.getSessionDetailFromLocalStorage();

    if (detail.userID && detail.expirationSeconds > 0) {
      this.dispatchSessionCreatedEvent(detail);
    }
  }
}
