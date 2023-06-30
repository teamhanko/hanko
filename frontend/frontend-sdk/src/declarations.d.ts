import {
  CustomEventWithDetail,
  SessionDetail,
  AuthFlowCompletedDetail,
  sessionCreatedType,
  sessionExpiredType,
  userLoggedOutType,
  userDeletedType,
  authFlowCompletedType,
} from "./lib/events/CustomEvents";

declare global {
  // eslint-disable-next-line no-unused-vars
  interface DocumentEventMap {
    [sessionCreatedType]: CustomEventWithDetail<SessionDetail>;
    [sessionExpiredType]: CustomEventWithDetail<null>;
    [userLoggedOutType]: CustomEventWithDetail<null>;
    [userDeletedType]: CustomEventWithDetail<null>;
    [authFlowCompletedType]: CustomEventWithDetail<AuthFlowCompletedDetail>;
  }
}

export {};
