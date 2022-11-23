/**
 * Every error thrown in the SDK is an instance of 'HankoError'. The value of the 'code' property is eligible to
 * translate the error into an error message.
 *
 * @extends {Error}
 * @category SDK
 * @subcategory Errors
 * @param code {string} - An error code that refers to the error instance.
 * @param cause {Error=} - The original error
 */
abstract class HankoError extends Error {
  code: string;
  cause?: Error;

  // eslint-disable-next-line require-jsdoc
  protected constructor(message: string, code: string, cause?: Error) {
    super(message);
    /**
     * @public
     * @type {string}
     */
    this.code = code;
    /**
     * @public
     * @type {Error=}
     */
    this.cause = cause;
    Object.setPrototypeOf(this, HankoError.prototype);
  }
}

/**
 * Every error that doesn't need to be handled in a special way is a 'TechnicalError'. Whenever you catch one, there is
 * usually nothing you can do but present an error to the user, e.g. "Something went wrong".
 *
 * @category SDK
 * @subcategory Errors
 * @extends {HankoError}
 */
class TechnicalError extends HankoError {
  // eslint-disable-next-line require-jsdoc
  constructor(cause?: Error) {
    super("Technical error", "somethingWentWrong", cause);
    Object.setPrototypeOf(this, TechnicalError.prototype);
  }
}

/**
 * Attempting to create a resource that already exists results in a 'ConflictError'.
 *
 * @category SDK
 * @subcategory Errors
 * @extends {HankoError}
 */
class ConflictError extends HankoError {
  // eslint-disable-next-line require-jsdoc
  constructor(userID?: string, cause?: Error) {
    super("Conflict error", "conflict", cause);
    Object.setPrototypeOf(this, ConflictError.prototype);
  }
}

/**
 * A 'RequestTimeoutError' occurs when the specified timeout has been reached.
 *
 * @category SDK
 * @subcategory Errors
 * @extends {HankoError}
 */
class RequestTimeoutError extends HankoError {
  // eslint-disable-next-line require-jsdoc
  constructor(cause?: Error) {
    super("Request timed out error", "requestTimeout", cause);
    Object.setPrototypeOf(this, RequestTimeoutError.prototype);
  }
}

/**
 * A 'WebauthnRequestCancelledError' occurs during WebAuthn authentication or registration, when the WebAuthn API throws
 * an error. In most cases, this happens when the user cancels the browser's WebAuthn dialog.
 *
 * @category SDK
 * @subcategory Errors
 * @extends {HankoError}
 */
class WebauthnRequestCancelledError extends HankoError {
  // eslint-disable-next-line require-jsdoc
  constructor(cause?: Error) {
    super("Request cancelled error", "requestCancelled", cause);
    Object.setPrototypeOf(this, WebauthnRequestCancelledError.prototype);
  }
}

/**
 * An 'InvalidPasswordError' occurs when invalid credentials are provided when logging in with a password.
 *
 * @category SDK
 * @subcategory Errors
 * @extends {HankoError}
 */
class InvalidPasswordError extends HankoError {
  // eslint-disable-next-line require-jsdoc
  constructor(cause?: Error) {
    super("Invalid password error", "invalidPassword", cause);
    Object.setPrototypeOf(this, InvalidPasswordError.prototype);
  }
}

/**
 * An 'InvalidPasswordError' occurs when an incorrect code is entered when logging in with a passcode.
 *
 * @category SDK
 * @subcategory Errors
 * @extends {HankoError}
 */
class InvalidPasscodeError extends HankoError {
  // eslint-disable-next-line require-jsdoc
  constructor(cause?: Error) {
    super("Invalid Passcode error", "invalidPasscode", cause);
    Object.setPrototypeOf(this, InvalidPasscodeError.prototype);
  }
}

/**
 * An 'InvalidWebauthnCredentialError' occurs if invalid credentials were used when logging in with WebAuthn.
 *
 * @category SDK
 * @subcategory Errors
 * @extends {HankoError}
 */
class InvalidWebauthnCredentialError extends HankoError {
  // eslint-disable-next-line require-jsdoc
  constructor(cause?: Error) {
    super(
      "Invalid WebAuthn credential error",
      "invalidWebauthnCredential",
      cause
    );
    Object.setPrototypeOf(this, InvalidWebauthnCredentialError.prototype);
  }
}

/**
 * A 'PasscodeExpiredError' occurs when the passcode has expired.
 *
 * @category SDK
 * @subcategory Errors
 * @extends {HankoError}
 */
class PasscodeExpiredError extends HankoError {
  // eslint-disable-next-line require-jsdoc
  constructor(cause?: Error) {
    super("Passcode expired error", "passcodeExpired", cause);
    Object.setPrototypeOf(this, PasscodeExpiredError.prototype);
  }
}

/**
 * A 'MaxNumOfPasscodeAttemptsReachedError' occurs when an incorrect passcode is provided too many times.
 *
 * @category SDK
 * @subcategory Errors
 * @extends {HankoError}
 */
class MaxNumOfPasscodeAttemptsReachedError extends HankoError {
  // eslint-disable-next-line require-jsdoc
  constructor(cause?: Error) {
    super(
      "Maximum number of Passcode attempts reached error",
      "passcodeAttemptsReached",
      cause
    );
    Object.setPrototypeOf(this, MaxNumOfPasscodeAttemptsReachedError.prototype);
  }
}

/**
 * A 'NotFoundError' occurs when the requested resource was not found.
 *
 * @category SDK
 * @subcategory Errors
 * @extends {HankoError}
 */
class NotFoundError extends HankoError {
  // eslint-disable-next-line require-jsdoc
  constructor(cause?: Error) {
    super("Not found error", "notFound", cause);
    Object.setPrototypeOf(this, NotFoundError.prototype);
  }
}

/**
 * A 'TooManyRequestsError' occurs due to rate limiting when too many requests are made.
 *
 * @category SDK
 * @subcategory Errors
 * @extends {HankoError}
 */
class TooManyRequestsError extends HankoError {
  retryAfter?: number;
  // eslint-disable-next-line require-jsdoc
  constructor(retryAfter?: number, cause?: Error) {
    super("Too many requests error", "tooManyRequests", cause);
    this.retryAfter = retryAfter;
    Object.setPrototypeOf(this, TooManyRequestsError.prototype);
  }
}

/**
 * An 'UnauthorizedError' occurs when the user is not authorized to access the resource.
 *
 * @category SDK
 * @subcategory Errors
 * @extends {HankoError}
 */
class UnauthorizedError extends HankoError {
  // eslint-disable-next-line require-jsdoc
  constructor(cause?: Error) {
    super("Unauthorized error", "unauthorized", cause);
    Object.setPrototypeOf(this, UnauthorizedError.prototype);
  }
}

/**
 * A 'UserVerificationError' occurs when the user verification requirements
 * for a WebAuthn ceremony are not met.
 *
 * @category SDK
 * @subcategory Errors
 * @extends {HankoError}
 */
class UserVerificationError extends HankoError {
  // eslint-disable-next-line require-jsdoc
  constructor(cause?: Error) {
    super("User verification error", "userVerification", cause);
    Object.setPrototypeOf(this, UserVerificationError.prototype);
  }
}

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
  UserVerificationError
};
