// SDK

import { Hanko } from "./Hanko";

export { Hanko };

// Clients

import { HttpClient } from "./lib/client/HttpClient";
import { Client } from "./lib/client/Client";
import { SessionClient } from "./lib/client/SessionClient";
import { UserClient } from "./lib/client/UserClient";

export { HttpClient, Client, SessionClient, UserClient };

// Events

import { Relay } from "./lib/events/Relay";

export { Relay };

// Utils

import { WebauthnSupport } from "./lib/WebauthnSupport";
import {
  generateCodeVerifier,
  setStoredCodeVerifier,
  getStoredCodeVerifier,
  clearStoredCodeVerifier,
} from "./lib/Pkce";

export {
  WebauthnSupport,
  generateCodeVerifier,
  setStoredCodeVerifier,
  getStoredCodeVerifier,
  clearStoredCodeVerifier,
};

// DTO

import {
  Email,
  Emails,
  Identity,
  SessionCheckResponse,
  Claims,
} from "./lib/Dto";

export type { Email, Emails, Identity, SessionCheckResponse, Claims };

// Errors

import {
  HankoError,
  ConflictError,
  ForbiddenError,
  EmailAddressAlreadyExistsError,
  InvalidPasswordError,
  InvalidPasscodeError,
  InvalidWebauthnCredentialError,
  MaxNumOfEmailAddressesReachedError,
  MaxNumOfPasscodeAttemptsReachedError,
  NotFoundError,
  PasscodeExpiredError,
  RequestTimeoutError,
  TechnicalError,
  ThirdPartyError,
  TooManyRequestsError,
  UnauthorizedError,
  UserVerificationError,
  WebauthnRequestCancelledError,
} from "./lib/Errors";

export {
  HankoError,
  ConflictError,
  ForbiddenError,
  EmailAddressAlreadyExistsError,
  InvalidPasswordError,
  InvalidPasscodeError,
  InvalidWebauthnCredentialError,
  MaxNumOfEmailAddressesReachedError,
  MaxNumOfPasscodeAttemptsReachedError,
  NotFoundError,
  PasscodeExpiredError,
  RequestTimeoutError,
  TechnicalError,
  ThirdPartyError,
  TooManyRequestsError,
  UnauthorizedError,
  UserVerificationError,
  WebauthnRequestCancelledError,
};

// Events

import {
  CustomEventWithDetail,
  SessionDetail,
  FlowDetail,
  sessionCreatedType,
  sessionExpiredType,
  userLoggedOutType,
  userDeletedType,
} from "./lib/events/CustomEvents";

export type { SessionDetail };
export type { FlowDetail };

export {
  sessionCreatedType,
  sessionExpiredType,
  userLoggedOutType,
  userDeletedType,
  CustomEventWithDetail,
};

// Misc

import { CookieSameSite } from "./lib/Cookie";

export type { CookieSameSite };

// Flow
export * from "./lib/flow-api/State";
export * from "./lib/flow-api/types/flow";
export * from "./lib/flow-api/types/flowError";
export * from "./lib/flow-api/types/payload";
export * from "./lib/flow-api/types/state";
export * from "./lib/flow-api/types/input";
