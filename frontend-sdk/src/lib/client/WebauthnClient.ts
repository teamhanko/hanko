import { WebauthnState } from "../state/WebauthnState";
import {
  InvalidWebauthnCredentialError,
  TechnicalError,
  UnauthorizedError,
  WebauthnRequestCancelledError,
} from "../Errors";
import {
  create as createWebauthnCredential,
  get as getWebauthnCredential,
  CredentialCreationOptionsJSON,
  CredentialRequestOptionsJSON,
  PublicKeyCredentialWithAssertionJSON,
} from "@github/webauthn-json";
import { Attestation, User, WebauthnFinalized } from "../Dto";
import { WebauthnSupport } from "../WebauthnSupport";
import { Client } from "./Client";

/**
 * A class that handles WebAuthn authentication and registration.
 *
 * @constructor
 * @category SDK
 * @subcategory Clients
 * @extends {Client}
 */
class WebauthnClient extends Client {
  private state: WebauthnState;

  // eslint-disable-next-line require-jsdoc
  constructor(api: string, timeout: number) {
    super(api, timeout);
    /**
     *  @private
     *  @type {WebauthnState}
     */
    this.state = new WebauthnState();
  }

  /**
   * Performs a WebAuthn authentication ceremony. When 'userID' is specified, the API provides a list of
   * allowed credentials and the browser is able to present a list of suitable credentials to the user.
   *
   * @param {string=} userID - The user's UUID.
   * @return {Promise<void>}
   * @throws {WebauthnRequestCancelledError}
   * @throws {InvalidWebauthnCredentialError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api#tag/WebAuthn/operation/webauthnLoginInit
   * @see https://docs.hanko.io/api#tag/WebAuthn/operation/webauthnLoginFinal
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
        .catch((e) => {
          reject(e);
        })
        .then((w: WebauthnFinalized) => {
          this.state.read().addCredential(w.user_id, w.credential_id).write();
          return resolve();
        })
        .catch((e) => {
          reject(e);
        });
    });
  }

  /**
   * Performs a WebAuthn registration ceremony.
   *
   * @return {Promise<void>}
   * @throws {WebauthnRequestCancelledError}
   * @throws {RequestTimeoutError}
   * @throws {UnauthorizedError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api#tag/WebAuthn/operation/webauthnRegInit
   * @see https://docs.hanko.io/api#tag/WebAuthn/operation/webauthnRegFinal
   * @see https://www.w3.org/TR/webauthn-2/#sctn-registering-a-new-credential
   */
  register(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.client
        .post("/webauthn/registration/initialize")
        .then((response) => {
          if (response.ok) {
            return response.json();
          } else if (response.status >= 400 && response.status <= 499) {
            throw new UnauthorizedError();
          }

          throw new TechnicalError();
        })
        .catch((e) => {
          reject(e);
        })
        .then((challenge: CredentialCreationOptionsJSON) => {
          return createWebauthnCredential(challenge);
        })
        .catch((e) => {
          reject(new WebauthnRequestCancelledError(e));
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
          } else if (response.status >= 400 && response.status <= 499) {
            throw new UnauthorizedError();
          }

          throw new TechnicalError();
        })
        .catch((e) => {
          reject(e);
        })
        .then((w: WebauthnFinalized) => {
          this.state.read().addCredential(w.user_id, w.credential_id).write();

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
      WebauthnSupport.isPlatformAuthenticatorAvailable()
        .then((supported) => {
          if (!user.webauthn_credentials || !user.webauthn_credentials.length) {
            return resolve(supported);
          }

          const matches = this.state
            .read()
            .matchCredentials(user.id, user.webauthn_credentials);

          return resolve(supported && !matches.length);
        })
        .catch((e) => {
          reject(new TechnicalError(e));
        });
    });
  }
}

export { WebauthnClient };
