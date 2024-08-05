import {
  CustomEventWithDetail,
  SessionDetail,
  sessionCreatedType,
  sessionExpiredType,
  userLoggedOutType,
  userDeletedType,
} from "./lib/events/CustomEvents";

declare global {
  // eslint-disable-next-line no-unused-vars
  interface DocumentEventMap {
    [sessionCreatedType]: CustomEventWithDetail<SessionDetail>;
    [sessionExpiredType]: CustomEventWithDetail<null>;
    [userLoggedOutType]: CustomEventWithDetail<null>;
    [userDeletedType]: CustomEventWithDetail<null>;
  }
}

export {};
