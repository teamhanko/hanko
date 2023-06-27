import { SessionDetail, sessionExpiredType } from "./CustomEvents";
import { Listener } from "./Listener";
import { Scheduler } from "./Scheduler";
import { Dispatcher } from "./Dispatcher";
import { Session } from "../Session";

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
  _session = new Session();

  // eslint-disable-next-line require-jsdoc
  constructor() {
    super();
    this.listenEventDependencies();
  }

  /**
   * Removes the scheduled "hanko-session-expired" event and re-schedules a new event with updated expirationSeconds, to
   * ensure the "hanko-session-expired" event won't be triggered too early.
   *
   * @private
   * @param {SessionDetail} detail - The event detail.
   */
  private scheduleSessionExpiredEvent = (detail: SessionDetail) => {
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
   * Handles the "storage" event in case the local storage entry, that contains the session detail has been changed by
   * another window. Depending on the new value of `expirationSeconds`, it either dispatches a "hanko-session-created"
   * or a "hanko-session-expired" event.
   *
   * @private
   * @param {StorageEvent} event - The storage event object.
   */
  private handleStorageEvent = (event: StorageEvent) => {
    if (event.key !== "hanko_session") return;

    const sessionDetail = this._session.get();

    if (!sessionDetail) {
      this.dispatchSessionExpiredEvent();
      return;
    }

    if (this._session.isAuthFlowCompleted()) {
      this.dispatchAuthFlowCompletedEvent({ userID: sessionDetail.userID });
      return;
    }

    this.dispatchSessionCreatedEvent(sessionDetail);
  };

  /**
   * Listens for events sent in the current browser.
   *
   * @private
   */
  private listenEventDependencies() {
    this._listener.onSessionCreated(this.scheduleSessionExpiredEvent);
    this._listener.onSessionExpired(this.cancelSessionExpiredEvent);
    this._listener.onUserDeleted(this.cancelSessionExpiredEvent);
    this._listener.onUserLoggedOut(this.cancelSessionExpiredEvent);

    // Handle cases, where the session has been changed by another window.
    window.addEventListener("storage", this.handleStorageEvent);

    const sessionDetail = this._session.get();

    if (sessionDetail) {
      this.scheduleSessionExpiredEvent(sessionDetail);
    }
  }
}
