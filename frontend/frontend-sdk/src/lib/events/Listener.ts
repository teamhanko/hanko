import { Throttle } from "../Throttle";
import {
  CustomEventWithDetail,
  SessionDetail,
  sessionCreatedType,
  sessionExpiredType,
  userDeletedType,
  userLoggedOutType,
} from "./CustomEvents";

/**
 * A callback function to be executed when an event is triggered.
 *
 * @alias CallbackFunc
 * @typedef {function} CallbackFunc
 * @memberOf Listener
 */
// eslint-disable-next-line no-unused-vars
type CallbackFunc<T> = (detail: T) => any;

/**
 * A wrapped callback function that will execute the original callback.
 *
 * @ignore
 * @param {T} event - The event object passed in the event.
 */
// eslint-disable-next-line no-unused-vars
type WrappedCallback<T> = (event: CustomEventWithDetail<T>) => void;

/**
 * A function returned when adding an event listener. The function can be called to remove the corresponding event
 * listener.
 *
 * @alias CleanupFunc
 * @typedef {function} CleanupFunc
 * @memberOf Listener
 */
type CleanupFunc = () => void;

/**
 * @interface
 * @ignore
 * @property {Function} callback - The function to be executed.
 * @property {boolean=} once - Whether the event listener should be removed after being called once.
 */
interface EventListenerParams<T> {
  callback: CallbackFunc<T>;
  once?: boolean;
}

/**
 * @interface
 * @ignore
 * @extends {EventListenerParams<T>}
 * @property {string} type - The type of the event.
 * @property {boolean=} throttle - Whether the event listener should be throttled.
 */
interface EventListenerWithTypeParams<T> extends EventListenerParams<T> {
  type: string;
  throttle?: boolean;
}

/**
 * A class to bind event listener for custom events.
 *
 * @category SDK
 * @subcategory Events
 */
export class Listener {
  public throttleLimit = 1000;
  _addEventListener = document.addEventListener.bind(document);
  _removeEventListener = document.removeEventListener.bind(document);
  _throttle = Throttle.throttle;

  /**
   * Wraps the given callback.
   *
   * @param callback
   * @param throttle
   * @private
   * @return {WrappedCallback}
   */
  private wrapCallback<T>(
    callback: CallbackFunc<T>,
    throttle: boolean,
  ): WrappedCallback<T> {
    // The function that will be called when the event is triggered.
    const wrappedCallback = (event: CustomEventWithDetail<T>) => {
      callback(event.detail);
    };

    // Throttle the listener if multiple SDK instances could trigger the same event at the same time,
    // but the callback function should only be executed once.
    if (throttle) {
      return this._throttle(wrappedCallback, this.throttleLimit, {
        leading: true,
        trailing: false,
      });
    }

    return wrappedCallback;
  }

  /**
   * Adds an event listener with the specified type, callback function, and options.
   *
   * @private
   * @param {EventListenerWithTypeParams<T>} params - The parameters for the event listener.
   * @returns {CleanupFunc} This function can be called to remove the event listener.
   */
  private addEventListenerWithType<T>({
    type,
    callback,
    once = false,
    throttle = false,
  }: EventListenerWithTypeParams<T>): CleanupFunc {
    const wrappedCallback = this.wrapCallback(callback, throttle);
    this._addEventListener(type, wrappedCallback, { once });
    return () => this._removeEventListener(type, wrappedCallback);
  }

  /**
   * Maps the parameters for an event listener to the `EventListenerWithTypeParams` interface.
   *
   * @static
   * @private
   * @param {string} type - The type of the event.
   * @param {EventListenerParams<T>} params - The parameters for the event listener.
   * @param {boolean} [throttle=false] - Whether the event listener should be throttled.
   * @returns {EventListenerWithTypeParams<T>}
   **/
  private static mapAddEventListenerParams<T>(
    type: string,
    { once, callback }: EventListenerParams<T>,
    throttle?: boolean,
  ): EventListenerWithTypeParams<T> {
    return {
      type,
      callback,
      once,
      throttle,
    };
  }

  /**
   * Adds an event listener with the specified type, callback function, and options.
   *
   * @private
   * @param {string} type - The type of the event.
   * @param {EventListenerParams<T>} params - The parameters for the event listener.
   * @param {boolean=} throttle - Whether the event listener should be throttled.
   * @returns {CleanupFunc} This function can be called to remove the event listener.
   */
  private addEventListener<T>(
    type: string,
    params: EventListenerParams<T>,
    throttle?: boolean,
  ) {
    return this.addEventListenerWithType(
      Listener.mapAddEventListenerParams(type, params, throttle),
    );
  }

  /**
   * Adds an event listener for "hanko-session-created" events. Will be triggered across all browser windows, when the user
   * logs in, or when the page has been loaded or refreshed and there is a valid session.
   *
   * @param {CallbackFunc<SessionDetail>} callback - The function to be called when the event is triggered.
   * @param {boolean=} once - Whether the event listener should be removed after being called once.
   * @returns {CleanupFunc} This function can be called to remove the event listener.
   */
  public onSessionCreated(
    callback: CallbackFunc<SessionDetail>,
    once?: boolean,
  ): CleanupFunc {
    return this.addEventListener(sessionCreatedType, { callback, once }, true);
  }

  /**
   * Adds an event listener for "hanko-session-expired" events. The event will be triggered across all browser windows
   * as soon as the current JWT expires or the user logs out. It also triggers, when the user deletes the account in
   * another window.
   *
   * @param {CallbackFunc<null>} callback - The function to be called when the event is triggered.
   * @param {boolean=} once - Whether the event listener should be removed after being called once.
   * @returns {CleanupFunc} This function can be called to remove the event listener.
   */
  public onSessionExpired(
    callback: CallbackFunc<null>,
    once?: boolean,
  ): CleanupFunc {
    return this.addEventListener(sessionExpiredType, { callback, once }, true);
  }

  /**
   * Adds an event listener for hanko-user-deleted events. The event triggers, when the user has deleted the account in
   * the browser window where the deletion happened.
   *
   * @param {CallbackFunc<null>} callback - The function to be called when the event is triggered.
   * @param {boolean=} once - Whether the event listener should be removed after being called once.
   * @returns {CleanupFunc} This function can be called to remove the event listener.
   */
  public onUserLoggedOut(
    callback: CallbackFunc<null>,
    once?: boolean,
  ): CleanupFunc {
    return this.addEventListener(userLoggedOutType, { callback, once });
  }

  /**
   * Adds an event listener for hanko-user-deleted events. The event triggers, when the user has deleted the account.
   *
   * @param {CallbackFunc<null>} callback - The function to be called when the event is triggered.
   * @param {boolean=} once - Whether the event listener should be removed after being called once.
   * @returns {CleanupFunc} This function can be called to remove the event listener.
   */
  public onUserDeleted(
    callback: CallbackFunc<null>,
    once?: boolean,
  ): CleanupFunc {
    return this.addEventListener(userDeletedType, { callback, once });
  }
}
