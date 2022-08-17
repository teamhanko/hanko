import Cookies from "js-cookie";

import {
  ConflictError,
  InvalidPasscodeError,
  InvalidPasswordError,
  InvalidWebauthnCredentialError,
  MaxNumOfPasscodeAttemptsReachedError,
  NotFoundError,
  RequestTimeoutError,
  TechnicalError,
  TooManyRequestsError,
  UnauthorizedError,
  WebauthnRequestCancelledError,
} from "./Errors";

import {
  Attestation,
  Config,
  Me,
  Passcode,
  User,
  UserInfo,
  WebauthnFinalized,
} from "./Dto";

import { PasscodeState, PasswordState, WebauthnState } from "./State";

import { WebauthnSupport } from "./WebauthnSupport";

import {
  create as createWebauthnCredential,
  get as getWebauthnCredential,
  CredentialCreationOptionsJSON,
  CredentialRequestOptionsJSON,
  PublicKeyCredentialWithAssertionJSON,
} from "@github/webauthn-json";

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

  constructor(api: string, timeout: number = 13000) {
    this.api = api;
    this.timeout = timeout;
  }

  _fetch(path: string, options: RequestInit) {
    const api = this.api;
    const url = api + path;
    const timeout = this.timeout;
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
        const authToken = xhr.getResponseHeader("X-Auth-Token");

        if (authToken) {
          const secure = !!api.match("^https://");
          Cookies.set(cookieName, authToken, { secure });
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

/**
 * A class to be extended by the other client classes.
 *
 * @category SDK
 * @subcategory Internal
 * @param {string} api - The URL of your Hanko API instance
 * @param {number=} timeout - The request timeout in milliseconds
 */
abstract class AbstractClient {
  protected client: HttpClient;

  constructor(api: string, timeout = 13000) {
    /**
     *  @protected
     *  @type {HttpClient}
     */
    this.client = new HttpClient(api, timeout);
  }
}

/**
 * A class for retrieving configurations from the API.
 *
 * @category SDK
 * @subcategory Clients
 * @extends {AbstractClient}
 */
class ConfigClient extends AbstractClient {
  /**
   * Retrieves the frontend configuration.
   * @return {Promise<Config>}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://teamhanko.github.io/hanko/#/.well-known/getConfig
   */
  get() {
    return new Promise<Config>((resolve, reject) => {
      this.client
        .get("/.well-known/config")
        .then((response) => {
          if (response.ok) {
            return resolve(response.json());
          }

          throw new TechnicalError();
        })
        .catch((e) => {
          reject(e);
        });
    });
  }
}

/**
 * A class to manage user information.
 *
 * @constructor
 * @category SDK
 * @subcategory Clients
 * @extends {AbstractClient}
 */
class UserClient extends AbstractClient {
  /**
   * Fetches basic information about the user by providing an email address. Can be used while the user is logged out
   * and is helpful in deciding which type of login to choose. For example, if the user's email is not verified, you may
   * want to log in with a passcode, or if no WebAuthN credentials are registered, you may not want to use WebAuthN.
   *
   * @param {string} email - The user's email address.
   * @return {Promise<UserInfo>}
   * @throws {NotFoundError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://teamhanko.github.io/hanko/#/authentication/getUserId
   */
  getInfo(email: string): Promise<UserInfo> {
    return new Promise<UserInfo>((resolve, reject) => {
      this.client
        .post("/user", { email })
        .then((response) => {
          if (response.ok) {
            return response.json();
          } else if (response.status === 404) {
            throw new NotFoundError();
          } else {
            throw new TechnicalError();
          }
        })
        .then((u: UserInfo) => resolve(u))
        .catch((e) => {
          reject(e);
        });
    });
  }

  /**
   * Creates a new user. Afterwards, the step should be to verify the email address via passcode. If a 'ConflictError'
   * occurred, you may want to prompt the user to log in.
   *
   * @param {string} email - The email address of the user to be created.
   * @return {Promise<User>}
   * @throws {ConflictError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://teamhanko.github.io/hanko/#/management/createUser
   */
  create(email: string): Promise<User> {
    return new Promise<User>((resolve, reject) => {
      this.client
        .post("/users", { email })
        .then((response) => {
          if (response.ok) {
            return resolve(response.json());
          } else if (response.status === 409) {
            throw new ConflictError();
          } else {
            throw new TechnicalError();
          }
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  /**
   * Fetches the current user.
   *
   * @return {Promise<User>}
   * @throws {UnauthorizedError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://teamhanko.github.io/hanko/#/authentication/IsUserAuthorized
   * @see https://teamhanko.github.io/hanko/#/management/listUser
   */
  getCurrent(): Promise<User> {
    return new Promise<User>((resolve, reject) =>
      this.client
        .get("/me")
        .then((response) => {
          if (response.ok) {
            return response.json();
          } else if (
            response.status === 400 ||
            response.status === 401 ||
            response.status === 404
          ) {
            throw new UnauthorizedError();
          } else {
            throw new TechnicalError();
          }
        })
        .then((me: Me) => {
          return this.client.get(`/users/${me.id}`);
        })
        .then((response) => {
          if (response.ok) {
            return resolve(response.json());
          } else if (
            response.status === 400 ||
            response.status === 401 ||
            response.status === 404
          ) {
            throw new UnauthorizedError();
          } else {
            throw new TechnicalError();
          }
        })
        .catch((e) => {
          reject(e);
        })
    );
  }
}

/**
 * A class that handles WebAuthN authentication and registration.
 *
 * @constructor
 * @category SDK
 * @subcategory Clients
 * @extends {AbstractClient}
 */
class WebauthnClient extends AbstractClient {
  private state: WebauthnState;
  support: WebauthnSupport;

  constructor(api: string, timeout: number) {
    super(api, timeout);
    /**
     *  @private
     *  @type {WebauthnState}
     */
    this.state = new WebauthnState();
    /**
     *  @public
     *  @type {WebauthnSupport}
     */
    this.support = new WebauthnSupport();
  }

  /**
   * Performs a WebAuthN authentication ceremony. When 'userID' is specified, the API provides a list of
   * allowed credentials and the browser is able to present a list of suitable credentials to the user.
   *
   * @param {string=} userID - The user's UUID.
   * @return {Promise<void>}
   * @throws {WebauthnRequestCancelledError}
   * @throws {InvalidWebauthnCredentialError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://teamhanko.github.io/hanko/#/authentication/webauthnLoginInit
   * @see https://teamhanko.github.io/hanko/#/authentication/webauthnLoginFinal
   * @see https://www.w3.org/TR/webauthn-2/#authentication-ceremony
   */
  login(userID?: string): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      this.client
        .post("/webauthn/login/initialize", { user_id: userID })
        .then((response) => {
          if (response.ok) {
            return response.json();
          }

          throw new TechnicalError();
        })
        .catch((e) => {
          reject(e);
        })
        .then((challenge: CredentialRequestOptionsJSON) => {
          return getWebauthnCredential(challenge);
        })
        .catch((e) => {
          throw new WebauthnRequestCancelledError(e);
        })
        .then((assertion: PublicKeyCredentialWithAssertionJSON) => {
          return this.client.post("/webauthn/login/finalize", assertion);
        })
        .then((response) => {
          if (response.ok) {
            return response.json();
          } else if (response.status === 400 || response.status === 401) {
            throw new InvalidWebauthnCredentialError();
          } else {
            throw new TechnicalError();
          }
        })
        .then((w: WebauthnFinalized) => {
          this.state.addCredentialID(w.user_id, w.credential_id);
          return resolve();
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  /**
   * Performs a WebAuthN registration ceremony.
   *
   * @return {Promise<void>}
   * @throws {WebauthnRequestCancelledError}
   * @throws {RequestTimeoutError}
   * @throws {UnauthorizedError}
   * @throws {TechnicalError}
   * @see https://teamhanko.github.io/hanko/#/authentication/webauthnRegInit
   * @see https://teamhanko.github.io/hanko/#/authentication/webauthnRegFinal
   * @see https://www.w3.org/TR/webauthn-2/#sctn-registering-a-new-credential
   */
  register(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.client
        .post("/webauthn/registration/initialize")
        .then((response) => {
          if (response.ok) {
            return response.json();
          }

          throw new TechnicalError();
        })
        .then((challenge: CredentialCreationOptionsJSON) => {
          return createWebauthnCredential(challenge);
        })
        .catch((e) => {
          throw new WebauthnRequestCancelledError(e);
        })
        .then((attestation: Attestation) => {
          // The generated PublicKeyCredentialWithAttestationJSON object does not align with the API. The list of
          // supported transports must be available under a different path.
          attestation.transports = attestation.response.transports;

          return this.client.post(
            "/webauthn/registration/finalize",
            attestation
          );
        })
        .then((response) => {
          if (response.ok) {
            return response.json();
          }

          throw new TechnicalError();
        })
        .then((w: WebauthnFinalized) => {
          this.state.addCredentialID(w.user_id, w.credential_id);
          return resolve();
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  /**
   * Determines whether a credential registration ceremony should be performed. Returns 'true' when a platform
   * authenticator is available and the user's credentials do not intersect with the credentials already known on the
   * current browser/device.
   *
   * @param {User} user - The user object.
   * @return {Promise<boolean>}
   * @throws {TechnicalError}
   */
  shouldRegister(user: User): Promise<boolean> {
    return new Promise<boolean>((resolve, reject) => {
      this.support
        .isPlatformAuthenticatorAvailable()
        .then((supported) => {
          if (!user.webauthn_credentials || !user.webauthn_credentials.length) {
            return resolve(supported);
          }

          const matches = this.state.matchCredentials(
            user.id,
            user.webauthn_credentials
          );

          return resolve(supported && !matches.length);
        })
        .catch((e) => {
          reject(new TechnicalError(e));
        });
    });
  }
}

/**
 * A class to handle passwords.
 *
 * @constructor
 * @category SDK
 * @subcategory Clients
 * @extends {AbstractClient}
 */
class PasswordClient extends AbstractClient {
  private state: PasswordState;

  constructor(api: string, timeout: number) {
    super(api, timeout);
    /**
     *  @private
     *  @type {PasswordState}
     */
    this.state = new PasswordState();
  }

  /**
   * Logs in a user with a password.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} password - The password.
   * @return {Promise<void>}
   * @throws {TooManyRequestsError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://teamhanko.github.io/hanko/#/authentication/passwordLogin
   */
  login(userID: string, password: string): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      this.client
        .post("/password/login", { user_id: userID, password })
        .then((response) => {
          if (response.ok) {
            return resolve();
          } else if (response.status === 401) {
            throw new InvalidPasswordError();
          } else if (response.status === 429) {
            const retryAfter = parseInt(
              response.headers.get("X-Retry-After") || "0",
              10
            );

            this.state.setRetryAfter(userID, retryAfter);

            throw new TooManyRequestsError(retryAfter);
          } else {
            throw new TechnicalError();
          }
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  /**
   * Updates a password.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} password - The new password.
   * @return {Promise<void>}
   * @throws {RequestTimeoutError}
   * @throws {UnauthorizedError}
   * @throws {TechnicalError}
   * @see https://teamhanko.github.io/hanko/#/authentication/password
   */
  update(userID: string, password: string): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      this.client
        .put("/password", { user_id: userID, password })
        .then((response) => {
          if (response.ok) {
            return resolve();
          }

          throw new TechnicalError();
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  /**
   * Returns the seconds until the rate limiting is actives.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getRetryAfter(userID: string) {
    return this.state.getRetryAfter(userID);
  }
}

/**
 * A class to handle passcodes.
 *
 * @constructor
 * @category SDK
 * @subcategory Clients
 * @extends {AbstractClient}
 */
class PasscodeClient extends AbstractClient {
  private state: PasscodeState;

  constructor(api: string, timeout: number) {
    super(api, timeout);
    /**
     *  @private
     *  @type {PasscodeState}
     */
    this.state = new PasscodeState();
  }

  /**
   * Causes the API to send a new passcode to the user's email address.
   *
   * @param {string} userID - The UUID of the user.
   * @return {Promise<Passcode>}
   * @throws {TooManyRequestsError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://teamhanko.github.io/hanko/#/authentication/passcodeInit
   */
  initialize(userID: string): Promise<Passcode> {
    return new Promise<Passcode>((resolve, reject) => {
      this.client
        .post("/passcode/login/initialize", { user_id: userID })
        .then((response) => {
          if (response.ok) {
            return response.json();
          } else if (response.status === 429) {
            const retryAfter = parseInt(
              response.headers.get("X-Retry-After") || "0",
              10
            );

            this.state.setResendAfter(userID, retryAfter);

            throw new TooManyRequestsError(retryAfter);
          } else {
            throw new TechnicalError();
          }
        })
        .then((passcode: Passcode) => {
          const ttl = passcode.ttl;

          this.state.setActiveID(userID, passcode.id);
          this.state.setTTL(userID, ttl);

          return resolve(passcode);
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  /**
   * Validates the passcode obtained from the email.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} code - The passcode digests.
   * @return {Promise<void>}
   * @throws {InvalidPasscodeError}
   * @throws {MaxNumOfPasscodeAttemptsReachedError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://teamhanko.github.io/hanko/#/authentication/passcodeFinal
   */
  finalize(userID: string, code: string): Promise<void> {
    const passcodeID = this.state.getActiveID(userID);

    return new Promise<void>((resolve, reject) => {
      this.client
        .post("/passcode/login/finalize", { id: passcodeID, code })
        .then((response) => {
          if (response.ok) {
            this.state.removeActive(userID);
            this.state.setResendAfter(userID, 0);

            return resolve();
          } else if (response.status === 401) {
            throw new InvalidPasscodeError();
          } else if (response.status === 404 || response.status === 410) {
            this.state.removeActive(userID);

            throw new MaxNumOfPasscodeAttemptsReachedError();
          } else {
            throw new TechnicalError();
          }
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  /**
   * Returns the seconds until the current passcode is active.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getTTL(userID: string) {
    return this.state.getTTL(userID);
  }

  /**
   * Returns the seconds until the rate limiting is active.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getResendAfter(userID: string) {
    return this.state.getResendAfter(userID);
  }
}

export {
  HttpClient,
  ConfigClient,
  UserClient,
  WebauthnClient,
  PasswordClient,
  PasscodeClient,
  Headers,
  Response,
};
