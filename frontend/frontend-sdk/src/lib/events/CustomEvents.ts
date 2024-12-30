import { Claims } from "../Dto";

/**
 * The type of the `hanko-session-created` event.
 * @typedef {string} sessionCreatedType
 * @memberOf Listener
 */
export const sessionCreatedType: "hanko-session-created" =
  "hanko-session-created";

/**
 * The type of the `hanko-session-expired` event.
 * @typedef {string} sessionExpiredType
 * @memberOf Listener
 */
export const sessionExpiredType: "hanko-session-expired" =
  "hanko-session-expired";

/**
 * The type of the `hanko-user-logged-out` event.
 * @typedef {string} userLoggedOutType
 * @memberOf Listener
 */
export const userLoggedOutType: "hanko-user-logged-out" =
  "hanko-user-logged-out";

/**
 * The type of the `hanko-user-deleted` event.
 * @typedef {string} userDeletedType
 * @memberOf Listener
 */
export const userDeletedType: "hanko-user-deleted" = "hanko-user-deleted";

/**
 * The type of the `hanko-user-logged-in` event.
 * @typedef {string} userLoggedInType
 * @memberOf Listener
 */
export const userLoggedInType: "hanko-user-logged-in" = "hanko-user-logged-in";

/**
 * The type of the `hanko-user-created` event.
 * @typedef {string} userCreatedType
 * @memberOf Listener
 */
export const userCreatedType: "hanko-user-created" = "hanko-user-created";

/**
 * The data passed in the `hanko-session-created` or `hanko-session-resumed` event.
 *
 * @interface
 * @category SDK
 * @subcategory Events
 * @property {number} expirationSeconds - This property is deprecated. The number of seconds until the JWT expires.
 * @property {Claims} claims - The JSON web token associated with the session. Only present when the Hanko-API allows the JWT to be accessible client-side.
 */
export interface SessionDetail {
  claims: Claims;
  expirationSeconds: number; // deprecated
}

/**
 * A custom event that includes a detail object.
 *
 * @category SDK
 * @subcategory Events
 * @extends CustomEvent
 * @ignore
 * @param {string} type - The type of the event.
 * @param {T} detail - The detail object to include in the event.
 */
export class CustomEventWithDetail<T> extends CustomEvent<T> {
  // eslint-disable-next-line require-jsdoc
  constructor(type: string, detail: T) {
    super(type, { detail });
  }
}
