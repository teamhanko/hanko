import Cookies from "js-cookie";
import { RequestTimeoutError, TechnicalError } from "../Errors";
import { SessionState } from "../state/session/SessionState";
import { PasscodeState } from "../state/users/PasscodeState";
import { Dispatcher } from "../events/Dispatcher";

/**
 * This class wraps an XMLHttpRequest to maintain compatibility with the fetch API.
 *
 * @category SDK
 * @subcategory Internal
 * @param {XMLHttpRequest} xhr - The request to be wrapped.
 * @see HttpClient
 */
class Headers {
  _xhr: XMLHttpRequest;

  // eslint-disable-next-line require-jsdoc
  constructor(xhr: XMLHttpRequest) {
    this._xhr = xhr;
  }

  /**
   * Returns the response header with the given name.
   *
   * @param {string} name
   * @return {string}
   */
  getResponseHeader(name: string) {
    return this._xhr.getResponseHeader(name);
  }
}

/**
 * This class wraps an XMLHttpRequest to maintain compatibility with the fetch API.
 *
 * @category SDK
 * @subcategory Internal
 * @param {XMLHttpRequest} xhr - The request to be wrapped.
 * @see HttpClient
 */
class Response {
  headers: Headers;
  ok: boolean;
  status: number;
  statusText: string;
  url: string;
  _decodedJSON: any;
  xhr: XMLHttpRequest;

  // eslint-disable-next-line require-jsdoc
  constructor(xhr: XMLHttpRequest) {
    /**
     *  @public
     *  @type {Headers}
     */
    this.headers = new Headers(xhr);
    /**
     *  @public
     *  @type {boolean}
     */
    this.ok = xhr.status >= 200 && xhr.status <= 299;
    /**
     *  @public
     *  @type {number}
     */
    this.status = xhr.status;
    /**
     *  @public
     *  @type {string}
     */
    this.statusText = xhr.statusText;
    /**
     *  @public
     *  @type {string}
     */
    this.url = xhr.responseURL;
    /**
     *  @private
     *  @type {XMLHttpRequest}
     */
    this.xhr = xhr;
  }

  /**
   * Returns the JSON decoded response.
   *
   * @return {any}
   */
  json() {
    if (!this._decodedJSON) {
      this._decodedJSON = JSON.parse(this.xhr.response);
    }
    return this._decodedJSON;
  }

  /**
   * Returns the value for Retry-After contained in the response header.
   *
   * @param {string} name - The name of the header field
   * @return {number}
   */
  parseNumericHeader(name: string): number {
    const result = parseInt(this.headers.getResponseHeader(name), 10);
    return isNaN(result) ? 0 : result;
  }
}

/**
 * Internally used for communication with the Hanko API. It also handles authorization tokens to enable authorized
 * requests.
 *
 * Currently, there is an issue with Safari and on iOS 15 devices where decoding a JSON response via the fetch API
 * breaks the user gesture and the user is not able to use the authenticator. Therefore, this class uses XMLHttpRequests
 * instead of the fetch API, but maintains compatibility by wrapping the XMLHttpRequests. So, if the issues are fixed,
 * we can easily return to the fetch API.
 *
 * @category SDK
 * @subcategory Internal
 * @param {string} api - The URL of your Hanko API instance
 * @param {number=} timeout - The request timeout in milliseconds
 */
class HttpClient {
  timeout: number;
  api: string;
  authCookieName = "hanko";
  sessionState: SessionState;
  passcodeState: PasscodeState;
  dispatcher: Dispatcher;

  // eslint-disable-next-line require-jsdoc
  constructor(api: string, timeout = 13000) {
    this.api = api;
    this.timeout = timeout;
    this.sessionState = new SessionState();
    this.passcodeState = new PasscodeState();
    this.dispatcher = new Dispatcher();
  }

  // eslint-disable-next-line require-jsdoc
  _fetch(path: string, options: RequestInit, xhr = new XMLHttpRequest()) {
    const url = this.api + path;
    const timeout = this.timeout;
    const bearerToken = this.getAuthCookie();

    return new Promise<Response>(function (resolve, reject) {
      xhr.open(options.method, url, true);
      xhr.setRequestHeader("Accept", "application/json");
      xhr.setRequestHeader("Content-Type", "application/json");

      if (bearerToken) {
        xhr.setRequestHeader("Authorization", `Bearer ${bearerToken}`);
      }

      xhr.timeout = timeout;
      xhr.withCredentials = true;
      xhr.onload = () => {
        const response = new Response(xhr);
        resolve(response);
      };

      xhr.onerror = () => {
        reject(new TechnicalError());
      };

      xhr.ontimeout = () => {
        reject(new RequestTimeoutError());
      };

      xhr.send(options.body ? options.body.toString() : null);
    });
  }

  /**
   * Returns the authentication token that was stored in the cookie.
   *
   * @return {string}
   * @return {string}
   */
  getAuthCookie(): string {
    return Cookies.get(this.authCookieName);
  }

  /**
   * Stores the authentication token to the cookie.
   *
   * @param {string} token - The authentication token to be stored.
   */
  _setAuthCookie(token: string) {
    const secure = !!this.api.match("^https://");
    Cookies.set(this.authCookieName, token, { secure });
  }

  /**
   * Removes the cookie used for authentication.
   */
  removeAuthCookie() {
    Cookies.remove(this.authCookieName);
  }

  processResponseHeadersOnLogin(userID: string, response: Response) {
    let jwt = "";
    let expirationSeconds = 0;

    response.xhr
      .getAllResponseHeaders()
      .split("\r\n")
      .forEach((h) => {
        const header = h.toLowerCase();
        if (header.startsWith("x-auth-token")) {
          jwt = response.headers.getResponseHeader("X-Auth-Token");
        } else if (header.startsWith("x-session-lifetime")) {
          expirationSeconds = parseInt(
            response.headers.getResponseHeader("X-Session-Lifetime"),
            10
          );
        }
      });

    this.passcodeState.read().reset(userID).write();

    if (expirationSeconds > 0) {
      this.sessionState.read();

      if (jwt) {
        this._setAuthCookie(jwt);
        this.sessionState.setJWT(jwt);
      }

      this.sessionState.setExpirationSeconds(expirationSeconds);
      this.sessionState.setUserID(userID);
      this.sessionState.write();
      this.dispatcher.dispatchSessionCreatedEvent({
        jwt,
        userID,
        expirationSeconds,
      });
    }
    return {
      jwt,
      userID,
      expirationSeconds,
    };
  }

  /**
   * Performs a GET request.
   *
   * @param {string} path - The path to the requested resource.
   * @return {Promise<Response>}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   */
  get(path: string) {
    return this._fetch(path, { method: "GET" });
  }

  /**
   * Performs a POST request.
   *
   * @param {string} path - The path to the requested resource.
   * @param {any=} body - The request body.
   * @return {Promise<Response>}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   */
  post(path: string, body?: any) {
    return this._fetch(path, {
      method: "POST",
      body: JSON.stringify(body),
    });
  }

  /**
   * Performs a PUT request.
   *
   * @param {string} path - The path to the requested resource.
   * @param {any=} body - The request body.
   * @return {Promise<Response>}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   */
  put(path: string, body?: any) {
    return this._fetch(path, {
      method: "PUT",
      body: JSON.stringify(body),
    });
  }

  /**
   * Performs a PATCH request.
   *
   * @param {string} path - The path to the requested resource.
   * @param {any=} body - The request body.
   * @return {Promise<Response>}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   */
  patch(path: string, body?: any) {
    return this._fetch(path, {
      method: "PATCH",
      body: JSON.stringify(body),
    });
  }

  /**
   * Performs a DELETE request.
   *
   * @param {string} path - The path to the requested resource.
   * @return {Promise<Response>}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   */
  delete(path: string) {
    return this._fetch(path, {
      method: "DELETE",
    });
  }
}

export { Headers, Response, HttpClient };
