// SDK

import { Hanko } from "./Hanko";

export { Hanko };

// Clients

import { Client } from "./lib/client/Client";
import { UserClient } from "./lib/client/UserClient";
import { EmailClient } from "./lib/client/EmailClient";
import { ThirdPartyClient } from "./lib/client/ThirdPartyClient";
import { TokenClient } from "./lib/client/TokenClient";
import { EnterpriseClient } from "./lib/client/EnterpriseClient";
import { SessionClient } from "./lib/client/SessionClient";

export {
  Client,
  UserClient,
  EmailClient,
  ThirdPartyClient,
  TokenClient,
  EnterpriseClient,
  SessionClient,
};

// Utils

import { WebauthnSupport } from "./lib/WebauthnSupport";

export { WebauthnSupport };

// DTO

import {
  PasswordConfig,
  EmailConfig,
  AccountConfig,
  Config,
  WebauthnFinalized,
  TokenFinalized,
  UserInfo,
  Me,
  Credential,
  User,
  UserCreated,
  Passcode,
  WebauthnTransports,
  Attestation,
  Email,
  Emails,
  WebauthnCredential,
  WebauthnCredentials,
  Identity,
  SessionCheckResponse,
} from "./lib/Dto";

export type {
  PasswordConfig,
  EmailConfig,
  AccountConfig,
  Config,
  WebauthnFinalized,
  TokenFinalized,
  UserInfo,
  Me,
  Credential,
  User,
  UserCreated,
  Passcode,
  WebauthnTransports,
  Attestation,
  Email,
  Emails,
  WebauthnCredential,
  WebauthnCredentials,
  Identity,
  SessionCheckResponse,
};

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

// import { Options, State, Action } from "./lib/flow-api/State";
// export type { Options };
// export { State, Action };

export * from "./lib/flow-api/State";
export * from "./lib/flow-api/types/flow";
export * from "./lib/flow-api/types/error";
export * from "./lib/flow-api/types/payload";

import { LoginMethod, MFAMethod, LastLogin } from "./lib/flow-api/types/payload";
export type { LoginMethod, MFAMethod, LastLogin };
// export * from "lib/flow-api/types/input";
