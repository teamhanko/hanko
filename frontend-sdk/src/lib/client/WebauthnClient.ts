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
  state: WebauthnState;
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
    this.state = new WebauthnState();
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
   */
  async login(
    userID?: string,
    useConditionalMediation?: boolean
  ): Promise<void> {
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

    this.state
      .read()
      .addCredential(finalizeResponse.user_id, finalizeResponse.credential_id)
      .write();

    return;
  }

  /**
   * Performs a WebAuthn registration ceremony.
   *
   * @return {Promise<void>}
   * @throws {WebauthnRequestCancelledError}
   * @throws {RequestTimeoutError}
   * @throws {UnauthorizedError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api/public#tag/WebAuthn/operation/webauthnRegInit
   * @see https://docs.hanko.io/api/public#tag/WebAuthn/operation/webauthnRegFinal
   * @see https://www.w3.org/TR/webauthn-2/#sctn-registering-a-new-credential
   */
  async register(): Promise<void> {
    const challengeResponse = await this.client.post(
      "/webauthn/registration/initialize"
    );

    if (challengeResponse.status >= 400 && challengeResponse.status <= 499) {
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

    if (
      attestationResponse.status >= 400 &&
      attestationResponse.status <= 499
    ) {
      throw new UnauthorizedError();
    }
    if (!attestationResponse.ok) {
      throw new TechnicalError();
    }

    const finalizeResponse: WebauthnFinalized = attestationResponse.json();
    this.state
      .read()
      .addCredential(finalizeResponse.user_id, finalizeResponse.credential_id)
      .write();

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

    const matches = this.state
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
