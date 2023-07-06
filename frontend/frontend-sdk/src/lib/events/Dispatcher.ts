import {
  SessionDetail,
  CustomEventWithDetail,
  AuthFlowCompletedDetail,
  sessionCreatedType,
  sessionExpiredType,
  userDeletedType,
  authFlowCompletedType,
  userLoggedOutType,
} from "./CustomEvents";
import { SessionState } from "../state/session/SessionState";

/**
 * Options for Dispatcher
 *
 * @category SDK
 * @subcategory Internal
 * @property {string} localStorageKey - The prefix / name of the local storage keys.
 */
interface DispatcherOptions {
  localStorageKey: string;
}

/**
 * A class that dispatches custom events.
 *
 * @category SDK
 * @subcategory Internal
 * @param {DispatcherOptions} options - The options that can be used
 */
export class Dispatcher {
  _dispatchEvent = document.dispatchEvent.bind(document);
  _sessionState: SessionState;

  // eslint-disable-next-line require-jsdoc
  constructor(options: DispatcherOptions) {
    this._sessionState = new SessionState({ ...options });
  }

  /**
   * Dispatches a custom event.
   *
   * @param {string} type
   * @param {T} detail
   * @private
   */
  private dispatch<T>(type: string, detail: T) {
    this._dispatchEvent(new CustomEventWithDetail(type, detail));
  }

  /**
   * Dispatches a "hanko-session-created" event to the document with the specified detail.
   *
   * @param {SessionDetail} detail - The event detail.
   */
  public dispatchSessionCreatedEvent(detail: SessionDetail) {
    this.dispatch(sessionCreatedType, detail);
  }

  /**
   * Dispatches a "hanko-session-expired" event to the document.
   */
  public dispatchSessionExpiredEvent() {
    this.dispatch(sessionExpiredType, null);
  }

  /**
   * Dispatches a "hanko-user-logged-out" event to the document.
   */
  public dispatchUserLoggedOutEvent() {
    this.dispatch(userLoggedOutType, null);
  }

  /**
   * Dispatches a "hanko-user-deleted" event to the document.
   */
  public dispatchUserDeletedEvent() {
    this.dispatch(userDeletedType, null);
  }

  /**
   * Dispatches a "hanko-auth-flow-completed" event to the document with the specified detail.
   *
   * @param {AuthFlowCompletedDetail} detail - The event detail.
   */
  public dispatchAuthFlowCompletedEvent(detail: AuthFlowCompletedDetail) {
    this._sessionState.read().setAuthFlowCompleted(true).write();
    this.dispatch(authFlowCompletedType, detail);
  }
}
