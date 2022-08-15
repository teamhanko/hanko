import { PublicKeyCredentialWithAttestationJSON } from "@github/webauthn-json";

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {boolean} enabled - Indicates passwords are enabled, so the API would accept login attempts using passwords.
 */
interface PasswordConfig {
  enabled: boolean;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {PasswordConfig} password - The password configuration.
 */
interface Config {
  password: PasswordConfig;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} credential_id - The ID of the credential that was used.
 * @property {string} user_id - The ID of the user that was used.
 */
interface WebauthnFinalized {
  credential_id: string;
  user_id: string;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} id - The UUID of the user.
 * @property {boolean} verified - Indicates the user's email address is verified.
 * @property {boolean} has_webauthn_credential - Indicates that the user has registered a WebAuthn credential in the past.
 */
interface UserInfo {
  id: string;
  verified: boolean;
  has_webauthn_credential: boolean;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} id - The UUID of the current user.
 * @ignore
 */
interface Me {
  id: string;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} id - The WebAuthN credential ID.
 */
interface Credential {
  id: string;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} id - The user's UUID.
 * @property {string} email - The user's email.
 * @property {Credential[]} webauthn_credentials - A list of credentials that have been registered.
 */
interface User {
  id: string;
  email: string;
  webauthn_credentials: Credential[];
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} id - The UUID of the passcode.
 * @property {number} ttl - How long the code is active in seconds.
 */
interface Passcode {
  id: string;
  ttl: number;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string[]} transports - A list of WebAuthN AuthenticatorTransport, e.g.: "usb", "internal",...
 * @ignore
 */
interface Attestation extends PublicKeyCredentialWithAttestationJSON {
  transports: string[];
}

export type {
  PasswordConfig,
  Config,
  WebauthnFinalized,
  Credential,
  UserInfo,
  Me,
  User,
  Passcode,
  Attestation,
};
