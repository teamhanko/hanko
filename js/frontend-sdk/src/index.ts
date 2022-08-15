// Hanko

import { Hanko } from "./Hanko";

export { Hanko };

// Client

import {
  ConfigClient,
  UserClient,
  WebauthnClient,
  PasswordClient,
  PasscodeClient,
} from "./lib/Client";

export {
  ConfigClient,
  UserClient,
  WebauthnClient,
  PasswordClient,
  PasscodeClient,
};

// WebauthnSupport

import { WebauthnSupport } from "./lib/WebauthnSupport";

export { WebauthnSupport };

// DTO

import {
  PasswordConfig,
  Config,
  WebauthnFinalized,
  Credential,
  UserInfo,
  User,
  Passcode,
} from "./lib/Dto";

export type {
  PasswordConfig,
  Config,
  WebauthnFinalized,
  Credential,
  UserInfo,
  User,
  Passcode,
};

// Errors

import {
  HankoError,
  TechnicalError,
  ConflictError,
  RequestTimeoutError,
  WebAuthnRequestCancelledError,
  InvalidPasswordError,
  InvalidPasscodeError,
  InvalidWebauthnCredentialError,
  PasscodeExpiredError,
  MaxNumOfPasscodeAttemptsReachedError,
  NotFoundError,
  TooManyRequestsError,
  UnauthorizedError,
} from "./lib/Errors";

export {
  HankoError,
  TechnicalError,
  ConflictError,
  RequestTimeoutError,
  WebAuthnRequestCancelledError,
  InvalidPasswordError,
  InvalidPasscodeError,
  InvalidWebauthnCredentialError,
  PasscodeExpiredError,
  MaxNumOfPasscodeAttemptsReachedError,
  NotFoundError,
  TooManyRequestsError,
  UnauthorizedError,
};
