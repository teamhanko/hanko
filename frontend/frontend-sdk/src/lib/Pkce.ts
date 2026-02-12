const PKCE_STORAGE_KEY = "hanko_pkce_code_verifier";

/**
 * Generates a high-entropy, cryptographically strong random string for use as a PKCE code verifier.
 * It uses rejection sampling to eliminate modulo bias, ensuring a perfectly uniform distribution
 * across the character set recommended by RFC 7636.
 *
 * @returns {string} A 64-character URL-safe random string.
 */
export const generateCodeVerifier = (): string => {
  const charset =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~";
  const n = charset.length;
  const maxValidByte = 256 - (256 % n);
  let result = "";
  const requiredLength = 64;
  const tempArray = new Uint8Array(1);

  while (result.length < requiredLength) {
    window.crypto.getRandomValues(tempArray);
    const byte = tempArray[0];
    if (byte < maxValidByte) {
      result += charset.charAt(byte % n);
    }
  }
  return result;
};

/**
 * Stores the PKCE code verifier in sessionStorage.
 * @param {string} verifier - The verifier to store.
 */
export const setStoredCodeVerifier = (verifier: string) => {
  if (typeof window !== "undefined" && window.sessionStorage) {
    window.sessionStorage.setItem(PKCE_STORAGE_KEY, verifier);
  }
};

/**
 * Retrieves the PKCE code verifier from sessionStorage.
 * @returns {string | null}
 */
export const getStoredCodeVerifier = (): string | null => {
  if (typeof window !== "undefined" && window.sessionStorage) {
    return window.sessionStorage.getItem(PKCE_STORAGE_KEY);
  }
  return null;
};

/**
 * Removes the PKCE code verifier from sessionStorage.
 */
export const clearStoredCodeVerifier = () => {
  if (typeof window !== "undefined" && window.sessionStorage) {
    window.sessionStorage.removeItem(PKCE_STORAGE_KEY);
  }
};
