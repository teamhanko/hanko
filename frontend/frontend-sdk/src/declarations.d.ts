import {
  CustomEventWithDetail,
  SessionCreatedEventDetail,
  AuthFlowCompletedEventDetail,
  sessionCreatedType,
  sessionRemovedType,
  userDeletedType,
  authFlowCompletedType,
} from "./lib/events/CustomEvents";

declare global {
  // eslint-disable-next-line no-unused-vars
  interface DocumentEventMap {
    [sessionCreatedType]: CustomEventWithDetail<SessionCreatedEventDetail>;
    [sessionRemovedType]: CustomEventWithDetail<null>;
    [userDeletedType]: CustomEventWithDetail<null>;
    [authFlowCompletedType]: CustomEventWithDetail<AuthFlowCompletedEventDetail>;
  }
}

export {};
