import JSCookie from "js-cookie";

/**
 * Options for Cookie
 *
 * @category SDK
 * @subcategory Internal
 * @property {string} cookieName - The name of the session cookie set from the SDK.
 */
interface CookieOptions {
  cookieName: string;
}

/**
 * A class to manage cookies.
 *
 * @category SDK
 * @subcategory Internal
 */
export class Cookie {
  authCookieName: string;

  // eslint-disable-next-line require-jsdoc
  constructor(options: CookieOptions) {
    this.authCookieName = options.cookieName;
  }

  /**
   * Returns the authentication token that was stored in the cookie.
   *
   * @return {string}
   */
  getAuthCookie(): string {
    return JSCookie.get(this.authCookieName);
  }

  /**
   * Stores the authentication token to the cookie.
   *
   * @param {string} token - The authentication token to be stored.
   * @param {boolean=} secure - Indicates a secure cookie should be set. Default is `true`.
   */
  setAuthCookie(token: string, secure = true) {
    JSCookie.set(this.authCookieName, token, { secure });
  }

  /**
   * Removes the cookie used for authentication.
   */
  removeAuthCookie() {
    JSCookie.remove(this.authCookieName);
  }
}
