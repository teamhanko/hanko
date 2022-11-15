// SDK

import { Hanko } from "./Hanko";

export { Hanko };

// Clients

import { ConfigClient } from "./lib/client/ConfigClient";
import { PasscodeClient } from "./lib/client/PasscodeClient";
import { PasswordClient } from "./lib/client/PasswordClient";
import { UserClient } from "./lib/client/UserClient";
import { WebauthnClient } from "./lib/client/WebauthnClient";

export {
  ConfigClient,
  UserClient,
  WebauthnClient,
  PasswordClient,
  PasscodeClient,
};

// Utils

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
  WebauthnRequestCancelledError,
  InvalidPasswordError,
  InvalidPasscodeError,
  InvalidWebauthnCredentialError,
  PasscodeExpiredError,
  MaxNumOfPasscodeAttemptsReachedError,
  NotFoundError,
  TooManyRequestsError,
  UnauthorizedError,
  UserVerificationError
} from "./lib/Errors";

export {
  HankoError,
  TechnicalError,
  ConflictError,
  RequestTimeoutError,
  WebauthnRequestCancelledError,
  InvalidPasswordError,
  InvalidPasscodeError,
  InvalidWebauthnCredentialError,
  PasscodeExpiredError,
  MaxNumOfPasscodeAttemptsReachedError,
  NotFoundError,
  TooManyRequestsError,
  UnauthorizedError,
  UserVerificationError,
};
