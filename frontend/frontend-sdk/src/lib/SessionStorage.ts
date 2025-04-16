/**
 * Options for SessionStorage
 *
 * @category SDK
 * @subcategory Internal
 * @property {string} keyName - The name of the sessionStorage session token entry set from the SDK.
 */
interface SessionStorageOptions {
  keyName: string;
}

/**
 * A class to manage sessionStorage.
 *
 * @category SDK
 * @subcategory Internal
 * @param {SessionStorageOptions} options - The options that can be used.
 */
export class SessionStorage {
  keyName: string;

  // eslint-disable-next-line require-jsdoc
  constructor(options: SessionStorageOptions) {
    this.keyName = options.keyName;
  }

  /**
   * Return the session token that was stored in the sessionStorage.
   *
   * @return {string}
   */
  getSessionToken(): string {
    return sessionStorage.getItem(this.keyName);
  }

  /**
   * Stores the session token in the sessionStorage.
   *
   * @param {string} token - The session token to be stored.
   */
  setSessionToken(token: string) {
    sessionStorage.setItem(this.keyName, token);
  }

  /**
   * Removes the session token used for authentication.
   */
  removeSessionToken() {
    sessionStorage.removeItem(this.keyName);
  }
}
