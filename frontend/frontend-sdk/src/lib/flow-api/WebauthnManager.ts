import {
  CredentialRequestOptionsJSON,
  CredentialCreationOptionsJSON,
  PublicKeyCredentialWithAssertionJSON,
  PublicKeyCredentialWithAttestationJSON,
  create,
  get,
} from "@github/webauthn-json";
import { RequestTimeoutError } from "../Errors";

// Applied when the creation options carry no usable `timeout`. Mirrors the
// Hanko API's own `webauthn.timeouts.registration` default, so the deadline
// enforced here matches the one the server would have asked for.
const DEFAULT_CREATION_TIMEOUT_MS = 600000;

/**
 * Manages WebAuthn credential operations as a singleton, ensuring only one active request at a time.
 * Uses an internal AbortController to cancel previous requests when a new one is initiated.
 */
class WebauthnManager {
  private static instance: WebauthnManager | null = null;
  private abortController = new AbortController();
  // eslint-disable-next-line no-useless-constructor,require-jsdoc
  private constructor() {}

  /**
   * Gets the singleton instance of WebauthnManager.
   * Creates a new instance if one doesn't exist, otherwise returns the existing one.
   * @returns {WebauthnManager} The singleton instance
   */
  public static getInstance(): WebauthnManager {
    if (!WebauthnManager.instance) {
      WebauthnManager.instance = new WebauthnManager();
    }
    return WebauthnManager.instance;
  }

  /**
   * Creates a new abort signal, aborting any ongoing WebAuthn request.
   * @private
   * @returns {AbortSignal} The new abort signal
   */
  private createAbortSignal(): AbortSignal {
    this.abortController.abort(); // Cancel any ongoing request
    this.abortController = new AbortController();
    return this.abortController.signal;
  }

  /**
   * Retrieves a WebAuthn credential using the provided options.
   * Aborts any previous request before starting a new one.
   * @param {CredentialRequestOptionsJSON} options - The options for credential retrieval
   * @returns {Promise<PublicKeyCredentialWithAssertionJSON>} A promise resolving to the retrieved credential
   * @throws {DOMException} If the WebAuthn request fails (e.g., aborted, not allowed)
   */
  public async getWebauthnCredential(
    options: CredentialRequestOptionsJSON,
  ): Promise<PublicKeyCredentialWithAssertionJSON> {
    return await get({
      ...options,
      signal: this.createAbortSignal(),
    });
  }

  /**
   * Retrieves a WebAuthn credential with conditional UI mediation.
   * Aborts any previous request before starting a new one.
   * @param {CredentialRequestOptionsJSON} publicKey - The public key options for conditional retrieval
   * @returns {Promise<PublicKeyCredentialWithAssertionJSON>} A promise resolving to the retrieved credential
   * @throws {DOMException} If the WebAuthn request fails (e.g., aborted, not allowed)
   */
  public async getConditionalWebauthnCredential(
    publicKey: CredentialRequestOptionsJSON["publicKey"],
  ): Promise<PublicKeyCredentialWithAssertionJSON> {
    return await get({
      publicKey,
      mediation: "conditional" as CredentialMediationRequirement,
      signal: this.createAbortSignal(),
    });
  }

  /**
   * Creates a new WebAuthn credential using the provided options.
   * Aborts any previous request before starting a new one.
   *
   * The ceremony is bounded by the `timeout` given in the creation options
   * (falling back to {@link DEFAULT_CREATION_TIMEOUT_MS}). Some authenticators
   * leave `navigator.credentials.create()` pending indefinitely and do not
   * honor the WebAuthn `timeout` themselves, which would keep the caller
   * waiting forever; enforcing the deadline here turns that into a regular
   * rejection the caller can recover from.
   * @param {CredentialCreationOptionsJSON} options - The options for credential creation
   * @returns {Promise<PublicKeyCredentialWithAttestationJSON>} A promise resolving to the created credential
   * @throws {DOMException} If the WebAuthn request fails (e.g., aborted, not allowed)
   * @throws {RequestTimeoutError} If the ceremony does not complete before the deadline
   */
  public async createWebauthnCredential(
    options: CredentialCreationOptionsJSON,
  ): Promise<PublicKeyCredentialWithAttestationJSON> {
    const signal = this.createAbortSignal();
    // createAbortSignal() has just installed a fresh controller; hold on to it
    // so the deadline below aborts this ceremony rather than a later one.
    const controller = this.abortController;
    // A missing or zero `timeout` is not a request to end the ceremony at once
    // - the WebAuthn timeout is a hint that clients clamp anyway - so fall back
    // to the default instead of aborting before the user can even respond.
    const requestedTimeout = options.publicKey?.timeout;
    const timeout =
      requestedTimeout > 0 ? requestedTimeout : DEFAULT_CREATION_TIMEOUT_MS;
    let deadline: ReturnType<typeof setTimeout>;

    try {
      return await Promise.race([
        create({ ...options, signal }),
        new Promise<never>((_resolve, reject) => {
          deadline = setTimeout(() => {
            controller.abort();
            reject(new RequestTimeoutError());
          }, timeout);
        }),
      ]);
    } finally {
      clearTimeout(deadline);
    }
  }
}

export default WebauthnManager;
