/**
 * A class to check the browser's WebAuthN support.
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
      return await window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable();
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
      typeof window.PublicKeyCredential.isExternalCTAP2SecurityKeySupported ===
        "function"
    ) {
      // @ts-ignore
      return await window.PublicKeyCredential.isExternalCTAP2SecurityKeySupported();
    }

    return this.supported();
  }
}

export { WebauthnSupport };
