import { RequestTimeoutError, TechnicalError } from "../Errors";
import { Dispatcher } from "../events/Dispatcher";
import { Cookie } from "../Cookie";
import { SessionStorage } from "../SessionStorage";
import { CookieAttributes } from "js-cookie";
import { HankoOptions } from "../../Hanko";

export type SessionTokenLocation = "cookie" | "sessionStorage";

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
   * Returns the response header value with the given `name` as a number. When the value is not a number the return
   * value will be 0.
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
 * Options for the HttpClient
 *
 * @category SDK
 * @subcategory Internal
 * @property {number=} timeout - The http request timeout in milliseconds.
 * @property {string} cookieName - The name of the session cookie set from the SDK.
 * @property {string=} cookieDomain - The domain where cookie set from the SDK is available. Defaults to the domain of the page where the cookie was created.
 * @property {string?} lang - The language used by the client(s) to convey to the Hanko API the language to use -
 *                           e.g. for translating outgoing emails - in a custom header (X-Language).
 * @property {string?} tenantId - The tenant ID for multi-tenant support, sent via X-Tenant-ID header.
 */
export interface HttpClientOptions {
  timeout?: number;
  cookieName?: string;
  cookieDomain?: string;
  lang?: string;
  sessionTokenLocation?: SessionTokenLocation;
  tenantId?: string;
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
 * @param {HttpClientOptions} options - The options the HttpClient must be provided
 */
class HttpClient {
  timeout: number;
  api: string;
  dispatcher: Dispatcher;
  cookie: Cookie;
  sessionTokenStorage: SessionStorage;
  lang: string;
  sessionTokenLocation: SessionTokenLocation;
  tenantId: string;

  // eslint-disable-next-line require-jsdoc
  constructor(api: string, options: HankoOptions) {
    this.api = api;
    this.timeout = options.timeout ?? 13000;
    this.dispatcher = new Dispatcher();
    this.cookie = new Cookie({ ...options });
    this.sessionTokenStorage = new SessionStorage({
      keyName: options.cookieName,
    });
    this.lang = options.lang;
    this.sessionTokenLocation = options.sessionTokenLocation;
    this.tenantId = options.tenantId;
  }

  // eslint-disable-next-line require-jsdoc
  _fetch(path: string, options: RequestInit, xhr = new XMLHttpRequest()) {
    const self = this;
    const url = this.api + path;
    const timeout = this.timeout;
    const bearerToken = this.getAuthToken();
    const lang = this.lang;
    const tenantId = this.tenantId;

    return new Promise<Response>(function (resolve, reject) {
      xhr.open(options.method, url, true);
      xhr.setRequestHeader("Accept", "application/json");
      xhr.setRequestHeader("Content-Type", "application/json");
      xhr.setRequestHeader("X-Language", lang);

      if (tenantId) {
        xhr.setRequestHeader("X-Tenant-ID", tenantId);
      }

      if (bearerToken) {
        xhr.setRequestHeader("Authorization", `Bearer ${bearerToken}`);
      }

      xhr.timeout = timeout;
      xhr.withCredentials = true;
      xhr.onload = () => {
        self.processHeaders(xhr);
        resolve(new Response(xhr));
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
   * Processes the response headers on login and extracts the JWT and expiration time.
   *
   * @param {XMLHttpRequest} xhr - The xhr object.
   */
  processHeaders(xhr: XMLHttpRequest) {
    let jwt = "";
    let expirationSeconds = 0;
    let retention = "";

    xhr
      .getAllResponseHeaders()
      .split("\r\n")
      .forEach((h) => {
        const header = h.toLowerCase();
        if (header.startsWith("x-auth-token")) {
          jwt = xhr.getResponseHeader("X-Auth-Token");
        } else if (header.startsWith("x-session-lifetime")) {
          expirationSeconds = parseInt(
            xhr.getResponseHeader("X-Session-Lifetime"),
            10,
          );
        } else if (header.startsWith("x-session-retention")) {
          retention = xhr.getResponseHeader("X-Session-Retention");
        }
      });

    if (jwt) {
      const https = new RegExp("^https://");
      const secure =
        !!this.api.match(https) && !!window.location.href.match(https);

      const expires =
        retention === "session"
          ? undefined
          : new Date(new Date().getTime() + expirationSeconds * 1000);

      this.setAuthToken(jwt, { secure, expires });
    }
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

  /**
   * Returns the session token either from the cookie or the sessionStorage.
   * @private
   * @return {string}
   */
  private getAuthToken(): string {
    let token = "";
    switch (this.sessionTokenLocation) {
      case "cookie":
        token = this.cookie.getAuthCookie();
        break;
      case "sessionStorage":
        token = this.sessionTokenStorage.getSessionToken();
        break;
      default:
        token = this.cookie.getAuthCookie();
        break;
    }
    return token;
  }

  /**
   * Stores the session token either in a cookie or in the sessionStorage depending on the configuration.
   * @param {string} token - The session token to be stored.
   * @param {CookieAttributes} options - Options for setting the auth cookie.
   * @private
   */
  private setAuthToken(token: string, options: CookieAttributes) {
    switch (this.sessionTokenLocation) {
      case "cookie":
        return this.cookie.setAuthCookie(token, options);
      case "sessionStorage":
        return this.sessionTokenStorage.setSessionToken(token);
      default:
        return this.cookie.setAuthCookie(token, options);
    }
  }
}

export { Headers, Response, HttpClient };
