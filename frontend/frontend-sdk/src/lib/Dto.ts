import { PublicKeyCredentialWithAttestationJSON } from "@github/webauthn-json";

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {boolean} enabled - Indicates passwords are enabled, so the API accepts login attempts using passwords.
 * @property {number} min_password_length - The minimum length of a password. To be used for password validation.
 */
export interface PasswordConfig {
  enabled: boolean;
  min_password_length: number;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {boolean} require_verification - Indicates that email addresses must be verified.
 * @property {number} max_num_of_addresses - The maximum number of email addresses a user can have.
 */
export interface EmailConfig {
  require_verification: boolean;
  max_num_of_addresses: number;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {boolean} allow_deletion - Indicates the current user is allowed to delete the account.
 * @property {boolean} allow_signup - Indicates the current user is allowed to sign up.
 */
export interface AccountConfig {
  allow_deletion: boolean;
  allow_signup: boolean;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {PasswordConfig} password - The password configuration.
 * @property {EmailConfig} emails - The email configuration.
 * @property {string[]} providers - The enabled third party providers.
 * @property {AccountConfig} account - Controls the behaviour regarding user accounts.
 * @property {boolean} use_enterprise - Enterprise Connections are enabled.
 */
export interface Config {
  password: PasswordConfig;
  emails: EmailConfig;
  providers: string[];
  account: AccountConfig;
  use_enterprise: boolean;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} credential_id - The ID of the credential that was used.
 * @property {string} user_id - The ID of the user that was used.
 */
export interface WebauthnFinalized {
  credential_id: string;
  user_id: string;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} user_id - The ID of the user associated with the token.
 */
export interface TokenFinalized {
  user_id: string;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} id - The UUID of the user.
 * @property {boolean} verified - Indicates whether the user's email address is verified.
 * @property {string} email_id - The UUID of the email address.
 * @property {boolean} has_webauthn_credential - Indicates that the user has registered a WebAuthn credential in the past.
 */
export interface UserInfo {
  id: string;
  verified: boolean;
  email_id: string;
  has_webauthn_credential: boolean;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} id - The UUID of the current user.
 * @ignore
 */
export interface Me {
  id: string;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} id - The WebAuthn credential ID.
 */
export interface Credential {
  id: string;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} id - The user's UUID.
 * @property {string=} email - The user's email.
 * @property {string=} username - The user's username.
 * @property {Credential[]} webauthn_credentials - A list of credentials that have been registered.
 */
export interface User {
  id: string;
  email?: string;
  username?: string;
  webauthn_credentials: Credential[];
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} user_id - The user's UUID.
 * @property {string} email_id - The ID of the new email.
 */
export interface UserCreated {
  user_id: string;
  email_id: string;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} id - The UUID of the passcode.
 * @property {number} ttl - How long the code is active in seconds.
 */
export interface Passcode {
  id: string;
  ttl: number;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string[]} - Transports which may be used by the authenticator. E.g. "internal", "ble",...
 */
export interface WebauthnTransports extends Array<string> {}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {WebauthnTransports} transports
 * @ignore
 */
export interface Attestation extends PublicKeyCredentialWithAttestationJSON {
  transports: WebauthnTransports;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} id - The UUID of the email address.
 * @property {string} address - The email address.
 * @property {boolean} is_verified - Indicates whether the email address is verified.
 * @property {boolean} is_primary - Indicates it's the primary email address.
 * @property {Identity} identity - Indicates that this email is linked to a third party account.
 * @property {Identity[]} identities - A list of identities, each identity indicates that this email is linked to a third party account.
 */
export interface Email {
  id: string;
  address: string;
  is_verified: boolean;
  is_primary: boolean;
  identity: Identity;
  identities: Identity[];
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {Email[]} - A list of emails assigned to the current user.
 */
export interface Emails extends Array<Email> {}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} id - The credential id.
 * @property {string=} name - The credential name.
 * @property {string} public_key - The public key.
 * @property {string} attestation_type - The attestation type.
 * @property {string} aaguid - The AAGUID of the authenticator.
 * @property {string} last_used_at - The time when the credential was used last.
 * @property {string} created_at - Time of credential creation.
 * @property {WebauthnTransports} transports
 */
export interface WebauthnCredential {
  id: string;
  name?: string;
  public_key: string;
  attestation_type: string;
  aaguid: string;
  last_used_at: string;
  created_at: string;
  transports: WebauthnTransports;
}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {WebauthnCredential[]} - A list of WebAuthn credential assigned to the current user.
 */
export interface WebauthnCredentials extends Array<WebauthnCredential> {}

/**
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} id - The subject ID with the third party provider.
 * @property {string} provider - The third party provider name.
 */
export interface Identity {
  id: string;
  provider: string;
}

/**
 * Represents the claims associated with a session or token. Includes standard claims such as `subject`, `issued_at`,
 * `expiration`, and others, as well as custom claims defined by the user.
 *
 * @template TCustomClaims - An optional generic parameter that represents custom claims.
 *                           It extends a record with string keys and unknown values.
 *                           Defaults to `Record<string, unknown>` if not provided.
 *
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {string} subject - The subject or identifier of the claims.
 * @property {string} [issued_at] - The timestamp when the claims were issued (optional).
 * @property {string} expiration - The timestamp when the claims expire.
 * @property {string[]} [audience] - The intended audience(s) for the claims (optional).
 * @property {string} [issuer] - The entity that issued the claims (optional).
 * @property {Pick<Email, "address" | "is_primary" | "is_verified">} [email] - Email information associated with the subject (optional).
 * @property {string} [username] - The subject's username (optional).
 * @property {string} session_id - The session identifier linked to the claims.
 *
 * @description Custom claims can be added via the `TCustomClaims` generic parameter, which will be merged
 * with the standard claims properties. These custom claims must follow the `Record<string, unknown>` pattern.
 */
export type Claims<
  TCustomClaims extends Record<string, unknown> = Record<string, unknown>,
> = {
  subject: string;
  issued_at?: string;
  expiration: string;
  audience?: string[];
  issuer?: string;
  email?: Pick<Email, "address" | "is_primary" | "is_verified">;
  username?: string;
  session_id: string;
} & TCustomClaims;

/**
 * Represents the response from a session validation or retrieval operation.
 *
 * @interface
 * @category SDK
 * @subcategory DTO
 * @property {boolean} is_valid - Indicates whether the session is valid.
 * @property {Claims} [claims] - The claims associated with the session (optional).
 * @property {string} [expiration_time] - The expiration timestamp of the session (optional).
 * @property {string} [user_id] - The user ID linked to the session (optional).
 */
export interface SessionCheckResponse {
  is_valid: boolean;
  claims?: Claims;
  expiration_time?: string;
  user_id?: string;
}
