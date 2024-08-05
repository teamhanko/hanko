// SDK

import { Hanko } from "./Hanko";

export { Hanko };

// Clients

import { UserClient } from "./lib/client/UserClient";
import { EmailClient } from "./lib/client/EmailClient";
import { ThirdPartyClient } from "./lib/client/ThirdPartyClient";
import { TokenClient } from "./lib/client/TokenClient";
import { EnterpriseClient } from "./lib/client/EnterpriseClient";

export {
  UserClient,
  EmailClient,
  ThirdPartyClient,
  TokenClient,
  EnterpriseClient,
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
  sessionCreatedType,
  sessionExpiredType,
  userLoggedOutType,
  userDeletedType,
} from "./lib/events/CustomEvents";

export type { SessionDetail };

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
