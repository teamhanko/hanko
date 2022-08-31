class HankoError extends Error {
  code: string;
  cause?: Error;
  constructor(message: string, code: string, cause?: Error) {
    super(message);
    this.code = code;
    this.cause = cause;
    Object.setPrototypeOf(this, HankoError.prototype);
  }
}

class TechnicalError extends HankoError {
  constructor(cause?: Error) {
    super("Technical error", "somethingWentWrong", cause);
    Object.setPrototypeOf(this, TechnicalError.prototype);
  }
}

class ConflictError extends HankoError {
  constructor(userID?: string, cause?: Error) {
    super("Conflict error", "conflict", cause);
    Object.setPrototypeOf(this, ConflictError.prototype);
  }
}

class RequestTimeoutError extends HankoError {
  constructor(cause?: Error) {
    super("Request timed out error", "requestTimeout", cause);
    Object.setPrototypeOf(this, RequestTimeoutError.prototype);
  }
}

class WebAuthnRequestCancelledError extends HankoError {
  constructor(cause?: Error) {
    super("Request cancelled error", "requestCancelled", cause);
    Object.setPrototypeOf(this, WebAuthnRequestCancelledError.prototype);
  }
}

class InvalidPasswordError extends HankoError {
  constructor(cause?: Error) {
    super("Invalid password error", "invalidPassword", cause);
    Object.setPrototypeOf(this, InvalidPasswordError.prototype);
  }
}

class InvalidPasscodeError extends HankoError {
  constructor(cause?: Error) {
    super("Invalid Passcode error", "invalidPasscode", cause);
    Object.setPrototypeOf(this, InvalidPasscodeError.prototype);
  }
}

class InvalidWebauthnCredentialError extends HankoError {
  constructor(cause?: Error) {
    super(
      "Invalid WebAuthn credential error",
      "invalidWebauthnCredential",
      cause
    );
    Object.setPrototypeOf(this, InvalidWebauthnCredentialError.prototype);
  }
}

class PasscodeExpiredError extends HankoError {
  constructor(cause?: Error) {
    super("Passcode expired error", "passcodeExpired", cause);
    Object.setPrototypeOf(this, PasscodeExpiredError.prototype);
  }
}

class MaxNumOfPasscodeAttemptsReachedError extends HankoError {
  constructor(cause?: Error) {
    super(
      "Maximum number of Passcode attempts reached error",
      "passcodeAttemptsReached",
      cause
    );
    Object.setPrototypeOf(this, MaxNumOfPasscodeAttemptsReachedError.prototype);
  }
}

class NotFoundError extends HankoError {
  constructor(cause?: Error) {
    super("Not found error", "notFound", cause);
    Object.setPrototypeOf(this, NotFoundError.prototype);
  }
}

class TooManyRequestsError extends HankoError {
  retryAfter?: number;
  constructor(retryAfter?: number, cause?: Error) {
    super("Too many requests error", "tooManyRequests", cause);
    this.retryAfter = retryAfter;
    Object.setPrototypeOf(this, TooManyRequestsError.prototype);
  }
}

class UnauthorizedError extends HankoError {
  constructor(cause?: Error) {
    super("Unauthorized error", "unauthorized", cause);
    Object.setPrototypeOf(this, UnauthorizedError.prototype);
  }
}

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
