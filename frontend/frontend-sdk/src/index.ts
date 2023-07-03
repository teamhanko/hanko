// SDK

import { Hanko } from "./Hanko";

export { Hanko };

// Clients

import { ConfigClient } from "./lib/client/ConfigClient";
import { PasscodeClient } from "./lib/client/PasscodeClient";
import { PasswordClient } from "./lib/client/PasswordClient";
import { UserClient } from "./lib/client/UserClient";
import { WebauthnClient } from "./lib/client/WebauthnClient";
import { EmailClient } from "./lib/client/EmailClient";
import { ThirdPartyClient } from "./lib/client/ThirdPartyClient";
import { TokenClient } from "./lib/client/TokenClient";

export {
  ConfigClient,
  UserClient,
  WebauthnClient,
  PasswordClient,
  PasscodeClient,
  EmailClient,
  ThirdPartyClient,
  TokenClient,
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
  AuthFlowCompletedDetail,
  authFlowCompletedType,
  sessionCreatedType,
  sessionExpiredType,
  userLoggedOutType,
  userDeletedType,
} from "./lib/events/CustomEvents";

export type { SessionDetail, AuthFlowCompletedDetail };

export {
  authFlowCompletedType,
  sessionCreatedType,
  sessionExpiredType,
  userLoggedOutType,
  userDeletedType,
  CustomEventWithDetail,
};
