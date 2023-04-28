/**
 * The type of the `hanko-session-created` event.
 * @typedef {string} sessionCreatedType
 * @memberOf Listener
 */
export const sessionCreatedType: "hanko-session-created" =
  "hanko-session-created";

/**
 * The type of the `hanko-session-removed` event.
 * @typedef {string} sessionRemovedType
 * @memberOf Listener
 */
export const sessionRemovedType: "hanko-session-removed" =
  "hanko-session-removed";

/**
 * The type of the `hanko-user-deleted` event.
 * @typedef {string} userDeletedType
 * @memberOf Listener
 */
export const userDeletedType: "hanko-user-deleted" = "hanko-user-deleted";

/**
 * The type of the `hanko-auth-flow-completed` event.
 * @typedef {string} authFlowCompletedType
 * @memberOf Listener
 */
export const authFlowCompletedType: "hanko-auth-flow-completed" =
  "hanko-auth-flow-completed";

/**
 * The data passed in the `hanko-session-created` event.
 *
 * @interface
 * @category SDK
 * @subcategory Events
 * @property {string=} jwt - The JSON web token associated with the session. Only present when the Hanko-API allows the JWT to be accessible client-side.
 * @property {number} expirationSeconds - The number of seconds until the JWT expires.
 * @property {string} userID - The user associated with the session.
 */
export interface SessionCreatedEventDetail {
  jwt?: string;
  expirationSeconds: number;
  userID: string;
}

/**
 * The data passed in the `hanko-auth-flow-completed` event.
 *
 * @interface
 * @category SDK
 * @subcategory Events
 * @property {string} userID - The user associated with the removed session.
 */
export interface AuthFlowCompletedEventDetail {
  userID: string;
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
