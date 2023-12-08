import JSCookie, { CookieAttributes } from "js-cookie";

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
 * @param {CookieOptions} options - The options that can be used
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
   * @param {CookieAttributes} options - Options for setting the auth cookie.
   */
  setAuthCookie(token: string, options: CookieAttributes = { secure: true }) {
    JSCookie.set(this.authCookieName, token, options);
  }

  /**
   * Removes the cookie used for authentication.
   */
  removeAuthCookie() {
    JSCookie.remove(this.authCookieName);
  }
}
