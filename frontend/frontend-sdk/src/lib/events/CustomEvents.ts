import { AllStates } from "../flow-api/types/flow";

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
 * The type of the `hanko-flow-state-changed` event.
 * @typedef {string} flowStateChangedType
 * @memberOf Listener
 */
export const flowStateChangedType: "hanko-flow-state-changed" =
  "hanko-flow-state-changed";

/**
 * The type of the `hanko-flow-error` event.
 * @typedef {string} flowErrorType
 * @memberOf Listener
 */
export const flowErrorType: "hanko-flow-error" = "hanko-flow-error";

/**
 * The data passed in the `hanko-session-created` or `hanko-session-resumed` event.
 *
 * @interface
 * @category SDK
 * @subcategory Events
 * @property {string=} jwt - The JSON web token associated with the session. Only present when the Hanko-API allows the JWT to be accessible client-side.
 * @property {number} expirationSeconds - The number of seconds until the JWT expires.
 */
export interface SessionDetail {
  jwt?: string;
  expirationSeconds: number;
}

export interface FlowDetail {
  state: AllStates;
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
