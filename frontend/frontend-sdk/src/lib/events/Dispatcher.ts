import {
  SessionCreatedEventDetail,
  CustomEventWithDetail,
  AuthFlowCompletedEventDetail,
  sessionCreatedType,
  sessionRemovedType,
  userDeletedType,
  authFlowCompletedType,
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
   * @param {SessionCreatedEventDetail} detail - The event detail.
   */
  public dispatchSessionCreatedEvent(detail: SessionCreatedEventDetail) {
    this.dispatch(sessionCreatedType, detail);
  }

  /**
   * Dispatches a "hanko-session-removed" event to the document.
   */
  public dispatchSessionRemovedEvent() {
    this.dispatch(sessionRemovedType, null);
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
   * @param {AuthFlowCompletedEventDetail} detail - The event detail.
   */
  public dispatchAuthFlowCompletedEvent(detail: AuthFlowCompletedEventDetail) {
    this.dispatch(authFlowCompletedType, detail);
  }
}
