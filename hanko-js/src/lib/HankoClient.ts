import {
  get as getWebauthnCredential,
  create as createWebauthnCredential,
  CredentialRequestOptionsJSON,
  PublicKeyCredentialWithAssertionJSON,
  CredentialCreationOptionsJSON,
  PublicKeyCredentialWithAttestationJSON,
} from "@github/webauthn-json";

import {
  PasscodeManager,
  PasswordManager,
  WebAuthnManager,
} from "./UserStateManager";

import {
  InvalidPasswordError,
  WebAuthnRequestCancelledError,
  NotFoundError,
  TooManyRequestsError,
  TechnicalError,
  MaxNumOfPasscodeAttemptsReachedError,
  InvalidPasscodeError,
  UnauthorizedError,
  EmailValidationRequiredError,
  InvalidWebauthnCredentialError,
  RequestTimeoutError,
} from "./Errors";

import { isUserVerifyingPlatformAuthenticatorAvailable } from "./WebauthnSupport";

export interface PasswordConfig {
  enabled: boolean;
}

export interface Config {
  password: PasswordConfig;
}

export interface WebauthnFinalized {
  credential_id: string;
  user_id: string;
}

export interface Credential {
  id: string;
}

export interface UserInfo {
  id: string;
  verified: boolean;
  has_webauthn_credential: boolean;
}

export interface User {
  id: string;
  email: string;
  webauthn_credentials: Credential[];
}

export interface Passcode {
  id: string;
  ttl: number;
}

interface Attestation extends PublicKeyCredentialWithAttestationJSON {
  transports: string[];
}

export class HankoClient {
  config: ConfigClient;
  user: UserClient;
  authenticator: WebauthnClient;
  password: PasswordClient;
  passcode: PasscodeClient;

  constructor(api: string, timeout: number) {
    this.config = new ConfigClient(api, timeout);
    this.user = new UserClient(api, timeout);
    this.authenticator = new WebauthnClient(api, timeout);
    this.password = new PasswordClient(api, timeout);
    this.passcode = new PasscodeClient(api, timeout);
  }
}

/** Class Response2 is part of a workaround for a Safari bug. See HttpClient._fetch2(). */
class Response2 {
  ok: boolean;
  status: number;
  statusText: string;
  url: string;
  decodedJSON: any;

  constructor(xhr: XMLHttpRequest) {
    this.ok = xhr.status >= 200 && xhr.status <= 299;
    this.status = xhr.status;
    this.statusText = xhr.statusText;
    this.url = xhr.responseURL;
    this.decodedJSON = JSON.parse(xhr.response);
  }

  json() {
    return this.decodedJSON;
  }
}

class HttpClient {
  timeout: number;
  api: string;

  constructor(api: string, timeout: number = 13000) {
    this.api = api;
    this.timeout = timeout;
  }

  _fetch(path: string, init: RequestInit) {
    return new Promise<Response>((resolve, reject) => {
      const controller = new AbortController();
      const timeout = setTimeout(() => controller.abort(), this.timeout);

      fetch(this.api + path, {
        mode: "cors",
        credentials: "include",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        signal: controller.signal,
        ...init,
      })
        .then((response) => {
          clearTimeout(timeout);

          return resolve(response);
        })
        .catch((e) => {
          reject(
            e.code === 20 ? new RequestTimeoutError(e) : new TechnicalError(e)
          );
        });
    });
  }

  /** _fetch2 a workaround for a Safari bug on iOS 15.5/15.6, which causes the WebAuthn actions to no longer
      be considered as user activated events, when using the fetch API. */
  _fetch2(method: string, path: string, body?: string) {
    const url = this.api + path;
    const timeout = this.timeout;

    return new Promise<Response2>(function (resolve, reject) {
      const xhr = new XMLHttpRequest();

      xhr.open(method, url, true);
      xhr.setRequestHeader("Accept", "application/json");
      xhr.setRequestHeader("Content-Type", "application/json");
      xhr.timeout = timeout;
      xhr.withCredentials = true;
      xhr.onload = () => {
        resolve(new Response2(xhr));
      };
      xhr.onerror = () => {
        reject(new TechnicalError());
      };
      xhr.ontimeout = () => {
        reject(new RequestTimeoutError());
      };
      xhr.send(body);
    });
  }

  get(path: string) {
    return this._fetch(path, { method: "GET" });
  }

  post(path: string, body?: any) {
    return this._fetch(path, {
      method: "POST",
      body: JSON.stringify(body),
    });
  }

  // post2 part of a workaround for a Safari bug. See _fetch2().
  post2(path: string, body?: any) {
    return this._fetch2("POST", path, JSON.stringify(body));
  }

  put(path: string, body?: any) {
    return this._fetch(path, {
      method: "PUT",
      body: JSON.stringify(body),
    });
  }
}

abstract class AbstractClient {
  client: HttpClient;

  constructor(api: string, timeout: number) {
    this.client = new HttpClient(api, timeout);
  }
}

class ConfigClient extends AbstractClient {
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

class UserClient extends AbstractClient {
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
        .then((u: UserInfo) => {
          if (!u.verified) {
            throw new EmailValidationRequiredError(u.id);
          }

          return resolve(u);
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  create(email: string): Promise<User> {
    return new Promise<User>((resolve, reject) => {
      this.client
        .post("/users", { email })
        .then((response) => {
          if (response.ok) {
            return resolve(response.json());
          } else if (response.status === 409) {
            throw new EmailValidationRequiredError();
          } else {
            throw new TechnicalError();
          }
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  getCurrent(): Promise<User> {
    return new Promise<User>((resolve, reject) =>
      this.client
        .get("/me")
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

class WebauthnClient extends AbstractClient {
  webAuthnManager: WebAuthnManager;

  constructor(api: string, timeout: number) {
    super(api, timeout);
    this.webAuthnManager = new WebAuthnManager();
  }

  login(userID?: string): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      this.client
        .post2("/webauthn/login/initialize", { user_id: userID })
        .then((response) => {
          if (response.ok) {
            return response.json();
          }

          throw new TechnicalError();
        })
        .then((challenge: CredentialRequestOptionsJSON) => {
          return getWebauthnCredential(challenge);
        })
        .catch((e) => {
          throw new WebAuthnRequestCancelledError(e);
        })
        .then((assertion: PublicKeyCredentialWithAssertionJSON) => {
          return this.client.post2("/webauthn/login/finalize", assertion);
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
          this.webAuthnManager.setCredentialID(w.user_id, w.credential_id);
          return resolve();
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  register(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.client
        .post2("/webauthn/registration/initialize")
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
          throw new WebAuthnRequestCancelledError(e);
        })
        .then((attestation: Attestation) => {
          // The generated PublicKeyCredentialWithAttestationJSON object does not align with the API. The list of
          // supported transports must be available under a different path.
          attestation.transports = attestation.response.transports;

          return this.client.post2(
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
          this.webAuthnManager.setCredentialID(w.user_id, w.credential_id);
          return resolve();
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  isAuthenticatorSupported() {
    return isUserVerifyingPlatformAuthenticatorAvailable();
  }

  shouldRegister(user: User): Promise<boolean> {
    return new Promise<boolean>((resolve, reject) => {
      this.isAuthenticatorSupported()
        .then((supported) => {
          if (!user.webauthn_credentials || !user.webauthn_credentials.length) {
            return resolve(supported);
          }

          const matches = this.webAuthnManager.matchCredentials(
            user.id,
            user.webauthn_credentials
          );

          return resolve(supported && !matches.length);
        })
        .catch((e) => {
          reject(e);
        });
    });
  }
}

class PasswordClient extends AbstractClient {
  passwordManager: PasswordManager;

  constructor(api: string, timeout: number) {
    super(api, timeout);
    this.passwordManager = new PasswordManager();
  }

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

            this.passwordManager.setRetryAfter(userID, retryAfter);

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

  getRetryAfter(userID: string) {
    return this.passwordManager.getRetryAfter(userID);
  }
}

class PasscodeClient extends AbstractClient {
  passcodeManager: PasscodeManager;

  constructor(api: string, timeout: number) {
    super(api, timeout);
    this.passcodeManager = new PasscodeManager();
  }

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

            this.passcodeManager.setResendAfter(userID, retryAfter);

            throw new TooManyRequestsError(retryAfter);
          } else {
            throw new TechnicalError();
          }
        })
        .then((passcode: Passcode) => {
          const ttl = passcode.ttl;

          this.passcodeManager.setActiveID(userID, passcode.id);
          this.passcodeManager.setTTL(userID, ttl);

          return resolve(passcode);
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  finalize = (userID: string, code: string): Promise<void> => {
    const passcodeID = this.passcodeManager.getActiveID(userID);

    return new Promise<void>((resolve, reject) => {
      this.client
        .post("/passcode/login/finalize", { id: passcodeID, code })
        .then((response) => {
          if (response.ok) {
            this.passcodeManager.removeActive(userID);
            this.passcodeManager.setResendAfter(userID, 0);

            return resolve();
          } else if (response.status === 401) {
            throw new InvalidPasscodeError();
          } else if (response.status === 404 || response.status === 410) {
            this.passcodeManager.removeActive(userID);

            throw new MaxNumOfPasscodeAttemptsReachedError();
          } else {
            throw new TechnicalError();
          }
        })
        .catch((e) => {
          reject(e);
        });
    });
  };

  getTTL(userID: string) {
    return this.passcodeManager.getTTL(userID);
  }

  getResendAfter(userID: string) {
    return this.passcodeManager.getResendAfter(userID);
  }
}

export default HankoClient;
