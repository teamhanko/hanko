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

/**
 * A class that dispatches custom events.
 *
 * @category SDK
 * @subcategory Internal
 */
export class Dispatcher {
  _dispatchEvent = document.dispatchEvent.bind(document);

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
    this.dispatch(authFlowCompletedType, detail);
  }
}
