import {
  isUserVerifyingPlatformAuthenticatorAvailable,
  get,
  create,
  CredentialRequestOptionsJSON,
  PublicKeyCredentialWithAssertionJSON,
  CredentialCreationOptionsJSON,
  PublicKeyCredentialWithAttestationJSON,
} from "@teamhanko/hanko-webauthn";

import { LocalStorage } from "./LocalStorage";

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

export class Hanko {
  config: ConfigClient;
  user: UserClient;
  authenticator: WebAuthnClient;
  password: PasswordClient;
  passcode: PasscodeClient;

  constructor(api: string) {
    this.config = new ConfigClient(api);
    this.user = new UserClient(api);
    this.authenticator = new WebAuthnClient(api);
    this.password = new PasswordClient(api);
    this.passcode = new PasscodeClient(api);
  }
}

class HttpClient {
  timeout: number;
  api: string;
  defaultHeaders: RequestInit = {
    mode: "cors",
    credentials: "include",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
  };

  constructor(api: string, timeout: number) {
    this.api = api;
    this.timeout = timeout;
  }

  _fetch(url: string, init: RequestInit) {
    return new Promise<Response>((resolve, reject) => {
      const controller = new AbortController();
      const timeout = setTimeout(() => controller.abort(), this.timeout);

      fetch(this.api + url, {
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

  get(url: string) {
    return this._fetch(url, { method: "GET" });
  }

  post(url: string, body?: any) {
    return this._fetch(url, {
      method: "POST",
      body: JSON.stringify(body),
    });
  }

  put(url: string, body?: any) {
    return this._fetch(url, {
      method: "PUT",
      body: JSON.stringify(body),
    });
  }
}

abstract class Utility {
  store: LocalStorage;
  client: HttpClient;

  constructor(api: string) {
    this.store = new LocalStorage("hanko");
    this.client = new HttpClient(api, 13000);
  }
}

class ConfigClient extends Utility {
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

class UserClient extends Utility {
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

class WebAuthnClient extends Utility {
  login(): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      this.client
        .post("/webauthn/login/initialize")
        .then((response) => {
          if (response.ok) {
            return response.json();
          }

          throw new TechnicalError();
        })
        .then((challenge: CredentialRequestOptionsJSON) => {
          return get(challenge);
        })
        .catch((e) => {
          throw new WebAuthnRequestCancelledError(e);
        })
        .then((assertion: PublicKeyCredentialWithAssertionJSON) => {
          return this.client.post("/webauthn/login/finalize", assertion);
        })
        .then((response) => {
          if (response.ok) {
            return response.json();
          } else if (response.status === 400) {
            throw new InvalidWebauthnCredentialError();
          } else {
            throw new TechnicalError();
          }
        })
        .then((w: WebauthnFinalized) => {
          this.store.setWebAuthnCredentialID(w.user_id, w.credential_id);
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
        .post("/webauthn/registration/initialize")
        .then((response) => {
          if (response.ok) {
            return response.json();
          }

          throw new TechnicalError();
        })
        .then((challenge: CredentialCreationOptionsJSON) => {
          return create(challenge);
        })
        .catch((e) => {
          throw new WebAuthnRequestCancelledError(e);
        })
        .then((attestation: PublicKeyCredentialWithAttestationJSON) => {
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
          this.store.setWebAuthnCredentialID(w.user_id, w.credential_id);
          return resolve();
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  isSupported() {
    return isUserVerifyingPlatformAuthenticatorAvailable();
  }

  shouldRegister(user: User): Promise<boolean> {
    return new Promise<boolean>((resolve, reject) => {
      this.isSupported()
        .then((supported) => {
          if (!user.webauthn_credentials) {
            return resolve(supported);
          }

          const hasCredentials = this.store.matchCredentials(
            user.id,
            user.webauthn_credentials
          );

          return resolve(supported && !hasCredentials);
        })
        .catch((e) => {
          reject(e);
        });
    });
  }
}

class PasswordClient extends Utility {
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

            this.store.setPasswordRetryAfter(userID, retryAfter);

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

  public getRetryAfter = (userID: string) => {
    return this.store.getPasswordRetryAfter(userID);
  };
}

class PasscodeClient extends Utility {
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

            this.store.setPasscodeRetryAfter(userID, retryAfter);

            throw new TooManyRequestsError(retryAfter);
          } else {
            throw new TechnicalError();
          }
        })
        .then((passcode: Passcode) => {
          const expiry = passcode.ttl;

          this.store.setActivePasscodeID(userID, passcode.id);
          this.store.setPasscodeExpiry(userID, expiry);

          return resolve(passcode);
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  finalize = (userID: string, code: string): Promise<void> => {
    const passcodeID = this.store.getActivePasscodeID(userID);

    return new Promise<void>((resolve, reject) => {
      this.client
        .post("/passcode/login/finalize", { id: passcodeID, code })
        .then((response) => {
          if (response.ok) {
            this.store.removeActivePasscodeID(userID);

            return resolve();
          } else if (response.status === 401) {
            throw new InvalidPasscodeError();
          } else if (response.status === 404 || response.status === 410) {
            this.store.removeActivePasscodeID(userID);
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

  getExpiry(userID: string) {
    return this.store.getPasscodeExpiry(userID);
  }

  getRetryAfter(userID: string) {
    return this.store.getPasscodeRetryAfter(userID);
  }
}

export default { HankoUtil: HttpClient };
