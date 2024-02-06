import JSCookie, { CookieAttributes } from "js-cookie";
import { TechnicalError } from "./Errors";

/**
 * Options for Cookie
 *
 * @category SDK
 * @subcategory Internal
 * @property {string} cookieName - The name of the session cookie set from the SDK.
 * @property {string=} cookieDomain - The domain where the cookie set from the SDK is available. Defaults to the domain of the page where the cookie was created.
 * @property {string=} cookieSameSite -Specify whether/when cookies are sent with cross-site requests. Defaults to "lax".
 */
interface CookieOptions {
  cookieName: string;
  cookieDomain?: string;
  cookieSameSite?: CookieSameSite;
}

export type CookieSameSite =
  | "strict"
  | "Strict"
  | "lax"
  | "Lax"
  | "none"
  | "None";

/**
 * A class to manage cookies.
 *
 * @category SDK
 * @subcategory Internal
 * @param {CookieOptions} options - The options that can be used
 */
export class Cookie {
  authCookieName: string;
  authCookieDomain?: string;
  authCookieSameSite: CookieSameSite;

  // eslint-disable-next-line require-jsdoc
  constructor(options: CookieOptions) {
    this.authCookieName = options.cookieName;
    this.authCookieDomain = options.cookieDomain;
    this.authCookieSameSite = options.cookieSameSite ?? "lax";
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
  setAuthCookie(token: string, options?: CookieAttributes) {
    const defaults: CookieAttributes = {
      secure: true,
      sameSite: this.authCookieSameSite,
    };

    if (this.authCookieDomain !== undefined) {
      defaults.domain = this.authCookieDomain;
    }

    const o: CookieAttributes = { ...defaults, ...options };

    if (
      (o.sameSite === "none" || o.sameSite === "None") &&
      o.secure === false
    ) {
      throw new TechnicalError(
        new Error("Secure attribute must be set when SameSite=None"),
      );
    }

    JSCookie.set(this.authCookieName, token, o);
  }

  /**
   * Removes the cookie used for authentication.
   */
  removeAuthCookie() {
    JSCookie.remove(this.authCookieName);
  }
}
