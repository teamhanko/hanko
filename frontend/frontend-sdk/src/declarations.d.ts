import {
  CustomEventWithDetail,
  SessionEventDetail,
  AuthFlowCompletedEventDetail,
  sessionCreatedType,
  sessionResumedType,
  sessionExpiredType,
  userLoggedOutType,
  userDeletedType,
  authFlowCompletedType,
} from "./lib/events/CustomEvents";

declare global {
  // eslint-disable-next-line no-unused-vars
  interface DocumentEventMap {
    [sessionCreatedType]: CustomEventWithDetail<SessionEventDetail>;
    [sessionResumedType]: CustomEventWithDetail<SessionEventDetail>;
    [sessionExpiredType]: CustomEventWithDetail<null>;
    [userLoggedOutType]: CustomEventWithDetail<null>;
    [userDeletedType]: CustomEventWithDetail<null>;
    [authFlowCompletedType]: CustomEventWithDetail<AuthFlowCompletedEventDetail>;
  }
}

export {};
