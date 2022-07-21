// This function does a simple check to test for the credential management API
// functions we need, and an indication of public key credential authentication
// support.
// https://developers.google.com/web/updates/2018/03/webauthn-credential-management
export function supported(): boolean {
  return !!(
    navigator.credentials &&
    navigator.credentials.create &&
    navigator.credentials.get &&
    window.PublicKeyCredential
  );
}

export async function isUserVerifyingPlatformAuthenticatorAvailable(): Promise<boolean> {
  if (
    supported() &&
    window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable
  ) {
    return await window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable();
  }

  return false;
}

export async function isSecurityKeySupported(): Promise<boolean> {
  if (
    window.PublicKeyCredential !== undefined &&
    // @ts-ignore
    typeof window.PublicKeyCredential.isExternalCTAP2SecurityKeySupported ===
      "function"
  ) {
    // @ts-ignore
    return await window.PublicKeyCredential.isExternalCTAP2SecurityKeySupported();
  }

  return supported();
}
