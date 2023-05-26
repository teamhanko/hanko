import {
  create as createWebauthnCredential,
  get as getWebauthnCredential,
} from "@github/webauthn-json";

import { WebauthnSupport } from "../WebauthnSupport";
import { Client } from "./Client";
import { PasscodeState } from "../state/users/PasscodeState";
import { WebauthnState } from "../state/users/WebauthnState";

import {
  InvalidWebauthnCredentialError,
  TechnicalError,
  UnauthorizedError,
  UserVerificationError,
  WebauthnRequestCancelledError,
} from "../Errors";

import {
  Attestation,
  User,
  WebauthnCredentials,
  WebauthnFinalized,
} from "../Dto";

/**
 * A class that handles WebAuthn authentication and registration.
 *
 * @constructor
 * @category SDK
 * @subcategory Clients
 * @extends {Client}
 */
class WebauthnClient extends Client {
  webauthnState: WebauthnState;
  passcodeState: PasscodeState;
  controller: AbortController;
  _getCredential = getWebauthnCredential;
  _createCredential = createWebauthnCredential;

  // eslint-disable-next-line require-jsdoc
  constructor(api: string, timeout = 13000) {
    super(api, timeout);
    /**
     *  @public
     *  @type {WebauthnState}
     */
    this.webauthnState = new WebauthnState();
    /**
     *  @public
     *  @type {PasscodeState}
     */
    this.passcodeState = new PasscodeState();
  }

  /**
   * Performs a WebAuthn authentication ceremony. When 'userID' is specified, the API provides a list of
   * allowed credentials and the browser is able to present a list of suitable credentials to the user.
   *
   * @param {string=} userID - The user's UUID.
   * @param {boolean=} useConditionalMediation - Enables autofill assisted login.
   * @return {Promise<void>}
   * @throws {WebauthnRequestCancelledError}
   * @throws {InvalidWebauthnCredentialError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api/public#tag/WebAuthn/operation/webauthnLoginInit
   * @see https://docs.hanko.io/api/public#tag/WebAuthn/operation/webauthnLoginFinal
   * @see https://www.w3.org/TR/webauthn-2/#authentication-ceremony
   * @return {WebauthnFinalized}
   */
  async login(
    userID?: string,
    useConditionalMediation?: boolean
  ): Promise<WebauthnFinalized> {
    const challengeResponse = await this.client.post(
      "/webauthn/login/initialize",
      { user_id: userID }
    );

    if (!challengeResponse.ok) {
      throw new TechnicalError();
    }

    const challenge = challengeResponse.json();
    challenge.signal = this._createAbortSignal();

    if (useConditionalMediation) {
      // `CredentialMediationRequirement` doesn't support "conditional" in the current typescript version.
      challenge.mediation = "conditional" as CredentialMediationRequirement;
    }

    let assertion;
    try {
      assertion = await this._getCredential(challenge);
    } catch (e) {
      throw new WebauthnRequestCancelledError(e);
    }

    const assertionResponse = await this.client.post(
      "/webauthn/login/finalize",
      assertion
    );

    if (assertionResponse.status === 400 || assertionResponse.status === 401) {
      throw new InvalidWebauthnCredentialError();
    } else if (!assertionResponse.ok) {
      throw new TechnicalError();
    }

    const finalizeResponse: WebauthnFinalized = assertionResponse.json();

    this.webauthnState
      .read()
      .addCredential(finalizeResponse.user_id, finalizeResponse.credential_id)
      .write();

    this.client.processResponseHeadersOnLogin(
      finalizeResponse.user_id,
      assertionResponse
    );

    return finalizeResponse;
  }

  /**
   * Performs a WebAuthn registration ceremony.
   *
   * @return {Promise<void>}
   * @throws {WebauthnRequestCancelledError}
   * @throws {RequestTimeoutError}
   * @throws {UnauthorizedError}
   * @throws {TechnicalError}
   * @throws {UserVerificationError}
   * @see https://docs.hanko.io/api/public#tag/WebAuthn/operation/webauthnRegInit
   * @see https://docs.hanko.io/api/public#tag/WebAuthn/operation/webauthnRegFinal
   * @see https://www.w3.org/TR/webauthn-2/#sctn-registering-a-new-credential
   */
  async register(): Promise<void> {
    const challengeResponse = await this.client.post(
      "/webauthn/registration/initialize"
    );

    if (challengeResponse.status === 401) {
      this.client.dispatcher.dispatchSessionExpiredEvent();
      throw new UnauthorizedError();
    } else if (!challengeResponse.ok) {
      throw new TechnicalError();
    }

    const challenge = challengeResponse.json();
    challenge.signal = this._createAbortSignal();

    let attestation;
    try {
      attestation = (await this._createCredential(challenge)) as Attestation;
    } catch (e) {
      throw new WebauthnRequestCancelledError(e);
    }

    // The generated PublicKeyCredentialWithAttestationJSON object does not align with the API. The list of
    // supported transports must be available under a different path.
    attestation.transports = attestation.response.transports;

    const attestationResponse = await this.client.post(
      "/webauthn/registration/finalize",
      attestation
    );

    if (attestationResponse.status === 401) {
      this.client.dispatcher.dispatchSessionExpiredEvent();
      throw new UnauthorizedError();
    }
    if (attestationResponse.status === 422) {
      throw new UserVerificationError();
    }
    if (!attestationResponse.ok) {
      throw new TechnicalError();
    }

    const finalizeResponse: WebauthnFinalized = attestationResponse.json();
    this.webauthnState
      .read()
      .addCredential(finalizeResponse.user_id, finalizeResponse.credential_id)
      .write();

    return;
  }

  /**
   * Returns a list of all WebAuthn credentials assigned to the current user.
   *
   * @return {Promise<WebauthnCredentials>}
   * @throws {UnauthorizedError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api/public#tag/WebAuthn/operation/listCredentials
   */
  async listCredentials(): Promise<WebauthnCredentials> {
    const response = await this.client.get("/webauthn/credentials");

    if (response.status === 401) {
      this.client.dispatcher.dispatchSessionExpiredEvent();
      throw new UnauthorizedError();
    } else if (!response.ok) {
      throw new TechnicalError();
    }

    return response.json();
  }

  /**
   * Updates the WebAuthn credential.
   *
   * @param {string=} credentialID - The credential's UUID.
   * @param {string} name - The new credential name.
   * @return {Promise<void>}
   * @throws {UnauthorizedError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api/public#tag/WebAuthn/operation/updateCredential
   */
  async updateCredential(credentialID: string, name: string): Promise<void> {
    const response = await this.client.patch(
      `/webauthn/credentials/${credentialID}`,
      {
        name,
      }
    );

    if (response.status === 401) {
      this.client.dispatcher.dispatchSessionExpiredEvent();
      throw new UnauthorizedError();
    } else if (!response.ok) {
      throw new TechnicalError();
    }

    return;
  }

  /**
   * Deletes the WebAuthn credential.
   *
   * @param {string=} credentialID - The credential's UUID.
   * @return {Promise<void>}
   * @throws {UnauthorizedError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api/public#tag/WebAuthn/operation/deleteCredential
   */
  async deleteCredential(credentialID: string): Promise<void> {
    const response = await this.client.delete(
      `/webauthn/credentials/${credentialID}`
    );

    if (response.status === 401) {
      this.client.dispatcher.dispatchSessionExpiredEvent();
      throw new UnauthorizedError();
    } else if (!response.ok) {
      throw new TechnicalError();
    }

    return;
  }

  /**
   * Determines whether a credential registration ceremony should be performed. Returns 'true' when WebAuthn
   * is supported and the user's credentials do not intersect with the credentials already known on the
   * current browser/device.
   *
   * @param {User} user - The user object.
   * @return {Promise<boolean>}
   */
  async shouldRegister(user: User): Promise<boolean> {
    const supported = WebauthnSupport.supported();

    if (!user.webauthn_credentials || !user.webauthn_credentials.length) {
      return supported;
    }

    const matches = this.webauthnState
      .read()
      .matchCredentials(user.id, user.webauthn_credentials);

    return supported && !matches.length;
  }

  // eslint-disable-next-line require-jsdoc
  _createAbortSignal() {
    if (this.controller) {
      this.controller.abort();
    }

    this.controller = new AbortController();
    return this.controller.signal;
  }
}

export { WebauthnClient };
