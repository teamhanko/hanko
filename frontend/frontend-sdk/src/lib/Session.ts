import { SessionDetail } from "./events/CustomEvents";
import { SessionState } from "./state/session/SessionState";
import { Cookie } from "./Cookie";

/**
 * Options for Session
 *
 * @category SDK
 * @subcategory Session
 * @property {string} cookieName - The name of the session cookie set from the SDK.
 * @property {string} localStorageKey - The prefix / name of the local storage keys.
 */
interface SessionOptions {
  cookieName: string;
  localStorageKey: string;
}

/**
 A class representing a session.

 @category SDK
 @subcategory Session
 @param {SessionOptions} options - The options that can be used
 */
export class Session {
  _sessionState: SessionState;
  _cookie: Cookie;

  // eslint-disable-next-line require-jsdoc
  constructor(options: SessionOptions) {
    this._sessionState = new SessionState({ ...options });
    this._cookie = new Cookie({ ...options });
  }

  /**
   Retrieves the session details.

   @returns {SessionDetail} The session details.
   */
  public get(): SessionDetail {
    const detail = this._get();
    return Session.validate(detail) ? detail : null;
  }

  /**
   Checks if the user is logged in.

   @returns {boolean} true if the user is logged in, false otherwise.
   */
  public isValid(): boolean {
    const session = this._get();
    return Session.validate(session);
  }

  /**
   Retrieves the session details.

   @ignore
   @returns {SessionDetail} The session details.
   */
  public _get(): SessionDetail {
    this._sessionState.read();

    const userID = this._sessionState.getUserID();
    const expirationSeconds = this._sessionState.getExpirationSeconds();
    const jwt = this._cookie.getAuthCookie();

    return {
      userID,
      expirationSeconds,
      jwt,
    };
  }

  /**
   Checks if the auth flow is completed. The value resets after the next login attempt.

   @returns {boolean} Returns true if the authentication flow is completed, false otherwise
   */
  public isAuthFlowCompleted() {
    this._sessionState.read();
    return this._sessionState.getAuthFlowCompleted();
  }

  /**
   Validates the session.

   @private
   @param {SessionDetail} detail - The session details to validate.
   @returns {boolean} true if the session details are valid, false otherwise.
   */
  private static validate(detail: SessionDetail): boolean {
    return !!(detail.expirationSeconds > 0 && detail.userID?.length);
  }
}
