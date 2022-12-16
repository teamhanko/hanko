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

export {
  ConfigClient,
  UserClient,
  WebauthnClient,
  PasswordClient,
  PasscodeClient,
  EmailClient,
  ThirdPartyClient,
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
  Email,
  Emails,
  WebauthnCredential,
  WebauthnCredentials,
  Passcode,
  Identity,
} from "./lib/Dto";

export type {
  PasswordConfig,
  Config,
  WebauthnFinalized,
  Credential,
  UserInfo,
  User,
  Email,
  Emails,
  WebauthnCredential,
  WebauthnCredentials,
  Passcode,
  Identity
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
  TooManyRequestsError,
  UnauthorizedError,
  UserVerificationError,
  WebauthnRequestCancelledError,
};
