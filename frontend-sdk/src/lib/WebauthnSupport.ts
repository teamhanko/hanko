/**
 * A class to check the browser's WebAuthn support.
 *
 * @hideconstructor
 * @category SDK
 * @subcategory Utilities
 */
class WebauthnSupport {
  /**
   * Does a simple check to test for the credential management API functions we need, and an indication of
   * public key credential authentication support.
   *
   * @see https://developers.google.com/web/updates/2018/03/webauthn-credential-management
   * @return boolean
   */
  static supported(): boolean {
    return !!(
      navigator.credentials &&
      navigator.credentials.create &&
      navigator.credentials.get &&
      window.PublicKeyCredential
    );
  }

  /**
   * Checks whether a user-verifying platform authenticator is available.
   *
   * @return Promise<boolean>
   */
  static async isPlatformAuthenticatorAvailable(): Promise<boolean> {
    if (
      this.supported() &&
      window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable
    ) {
      return window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable();
    }

    return false;
  }

  /**
   * Checks whether external CTAP2 security keys are supported.
   *
   * @return Promise<boolean>
   */
  static async isSecurityKeySupported(): Promise<boolean> {
    if (
      window.PublicKeyCredential !== undefined &&
      // @ts-ignore
      window.PublicKeyCredential.isExternalCTAP2SecurityKeySupported
    ) {
      // @ts-ignore
      return window.PublicKeyCredential.isExternalCTAP2SecurityKeySupported();
    }

    return this.supported();
  }

  /**
   * Checks whether autofill assisted requests are supported.
   *
   * @return Promise<boolean>
   */
  static async isConditionalMediationAvailable(): Promise<boolean> {
    if (
      // @ts-ignore
      window.PublicKeyCredential &&
      // @ts-ignore
      window.PublicKeyCredential.isConditionalMediationAvailable
    ) {
      // @ts-ignore
      return window.PublicKeyCredential.isConditionalMediationAvailable();
    }

    return false;
  }
}

export { WebauthnSupport };
