import Cookies from "js-cookie";
import { RequestTimeoutError, TechnicalError } from "../Errors";

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

  get(name: string) {
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
    this._decodedJSON = JSON.parse(xhr.response);
  }

  /**
   * Returns the JSON decoded response.
   *
   * @return {any}
   */
  json() {
    return this._decodedJSON;
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
  isSameOrigin: boolean;

  constructor(api: string, timeout = 13000) {
    this.api = api;
    this.timeout = timeout;
    this.isSameOrigin = this._detectSameOrigin(api);
  }

  _detectSameOrigin(api: string) {
    const loc = window.location;
    const a = document.createElement("a");

    a.href = api;

    if (
      a.hostname === "localhost" &&
      loc.hostname === "localhost" &&
      a.protocol === loc.protocol
    ) {
      return true;
    }

    return (
      a.hostname === loc.hostname &&
      a.port === loc.port &&
      a.protocol === loc.protocol
    );
  }

  _fetch(path: string, options: RequestInit) {
    const api = this.api;
    const url = api + path;
    const timeout = this.timeout;
    const isSameOrigin = this.isSameOrigin;
    const cookieName = "hanko";
    const bearerToken = Cookies.get(cookieName);

    return new Promise<Response>(function (resolve, reject) {
      const xhr = new XMLHttpRequest();

      xhr.open(options.method, url, true);
      xhr.setRequestHeader("Accept", "application/json");
      xhr.setRequestHeader("Content-Type", "application/json");

      if (bearerToken) {
        xhr.setRequestHeader("Authorization", `Bearer ${bearerToken}`);
      }

      xhr.timeout = timeout;
      xhr.withCredentials = true;
      xhr.onload = () => {
        if (!isSameOrigin) {
          const authToken = xhr.getResponseHeader("X-Auth-Token");

          if (authToken) {
            const secure = !!api.match("^https://");
            Cookies.set(cookieName, authToken, { secure });
          }
        }

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
}

export { Headers, Response, HttpClient };
