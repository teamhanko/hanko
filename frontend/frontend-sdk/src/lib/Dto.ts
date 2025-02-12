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
