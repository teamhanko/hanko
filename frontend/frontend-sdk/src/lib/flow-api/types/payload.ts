import {
  CredentialCreationOptionsJSON,
  CredentialRequestOptionsJSON,
} from "@github/webauthn-json/src/webauthn-json/basic/json";
import { Claims } from "../../Dto";

/**
 * Payload of the `passcode_confirmation` state, describing the passcode that was sent to the user.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {boolean} passcode_resent - Indicates whether the passcode was just resent.
 * @property {number} resend_after - The number of seconds to wait before the passcode can be resent again.
 */
export interface PasscodeConfirmationPayload {
  readonly passcode_resent: boolean;
  readonly resend_after: number;
}

/**
 * Payload of the `login_passkey` state, carrying the options needed to perform the WebAuthn assertion ceremony.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {CredentialRequestOptionsJSON} request_options - The WebAuthn request options to pass to `navigator.credentials.get()`.
 */
export interface LoginPasskeyPayload {
  readonly request_options: CredentialRequestOptionsJSON;
}

/**
 * Payload of the `mfa_otp_secret_creation` state, carrying the newly generated TOTP secret.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {string} otp_secret - The generated TOTP secret.
 * @property {string} otp_image_source - A data URL of a QR code image encoding the secret, for scanning with an authenticator app.
 */
export interface MFAOTPSecretCreationPayload {
  readonly otp_secret: string;
  readonly otp_image_source: string;
}

/**
 * Payload of states that expect a WebAuthn attestation (registration) ceremony to be performed,
 * carrying the options needed to create the credential.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {CredentialCreationOptionsJSON} creation_options - The WebAuthn creation options to pass to `navigator.credentials.create()`.
 */
export interface OnboardingVerifyPasskeyAttestationPayload {
  readonly creation_options: CredentialCreationOptionsJSON;
}

/**
 * Payload of the `login_init` state. `request_options` is present when passkey autofill / conditional
 * mediation should be attempted immediately.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {CredentialRequestOptionsJSON} [request_options] - The WebAuthn request options for conditional mediation, if applicable.
 */
export interface LoginInitPayload {
  readonly request_options?: CredentialRequestOptionsJSON;
}

/**
 * A WebAuthn credential (passkey or security key) registered to a user.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {string} id - The credential's unique ID.
 * @property {string} [name] - A user-assigned name for the credential.
 * @property {string} public_key - The credential's public key.
 * @property {string} attestation_type - The WebAuthn attestation type used during registration.
 * @property {string} aaguid - The authenticator's AAGUID.
 * @property {string} [last_used_at] - Timestamp of the last time the credential was used, if any.
 * @property {string} created_at - Timestamp of when the credential was created.
 * @property {string} transports - The transports supported by the authenticator (e.g. `usb`, `internal`).
 * @property {string} backup_eligible - Whether the credential is eligible for backup (e.g. via a passkey provider).
 * @property {string} backup_state - The credential's current backup state.
 */
export interface WebauthnCredential {
  readonly id: string;
  readonly name?: string;
  readonly public_key: string;
  readonly attestation_type: string;
  readonly aaguid: string;
  readonly last_used_at?: string;
  readonly created_at: string;
  readonly transports: string;
  readonly backup_eligible: string;
  readonly backup_state: string;
}

/**
 * A username assigned to a user.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {string} id - The username's unique ID.
 * @property {string} username - The username value.
 * @property {string} created_at - Timestamp of when the username was created.
 * @property {string} updated_at - Timestamp of when the username was last updated.
 */
export interface Username {
  id: string;
  username: string;
  created_at: string;
  updated_at: string;
}

/**
 * A third-party identity linked to a user, as seen through the flow API.
 *
 * Note: a same-named `Identity` interface also exists in `Dto.ts` for the standalone
 * user-management REST API; it happens to have the same shape today, but the two are declared
 * independently and are not guaranteed to stay in sync.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {string} id - The subject ID with the third-party provider.
 * @property {string} provider - The third-party provider name.
 * @property {string} [identity_id] - The ID of the identity link itself.
 */
// Known docs limitation: this collides by name with `Dto.ts`'s `Identity`. JSDoc has no default
// file/module scoping, so both land on the same generated Identity.html page as two separate
// "Properties" tables. See the note there too.
export interface Identity {
  readonly id: string;
  readonly provider: string;
  readonly identity_id?: string;
}

/**
 * An email address belonging to a user, as seen through the flow API.
 *
 * Note: a same-named `Email` interface also exists in `Dto.ts` for the standalone
 * user-management REST API; the two are declared independently and have slightly different shapes.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {string} id - The UUID of the email address.
 * @property {string} address - The email address.
 * @property {boolean} is_verified - Indicates whether the email address is verified.
 * @property {boolean} is_primary - Indicates it's the primary email address.
 * @property {Identity[]} [identities] - Third-party identities linked to this email address, if any.
 */
// Known docs limitation: this collides by name with `Dto.ts`'s `Email`. JSDoc has no default
// file/module scoping, so both land on the same generated Email.html page as two separate
// "Properties" tables. See the note there too.
export interface Email {
  readonly id: string;
  readonly address: string;
  readonly is_verified: boolean;
  readonly is_primary: boolean;
  readonly identities?: Identity[];
}

/**
 * A user's multi-factor authentication configuration.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {boolean} auth_app_set_up - Whether an authenticator app (TOTP) has been set up.
 * @property {boolean} totp_enabled - Whether TOTP is enabled as a second factor.
 * @property {boolean} security_keys_enabled - Whether security keys are enabled as a second factor.
 */
export interface MFAConfig {
  readonly auth_app_set_up: boolean;
  readonly totp_enabled: boolean;
  readonly security_keys_enabled: boolean;
}

/**
 * A user's custom metadata, split into a `public_metadata` portion (readable by any client) and an
 * `unsafe_metadata` portion (writable by any authenticated client).
 * @template PublicMetadata - The shape of the public metadata.
 * @template UnsafeMetadata - The shape of the unsafe metadata.
 * @category SDK
 * @subcategory Flow Payloads
 * @property {PublicMetadata} [public_metadata] - Metadata visible to any client.
 * @property {UnsafeMetadata} [unsafe_metadata] - Metadata that any authenticated client may write.
 */
export type UserMetadata<
  PublicMetadata extends Record<string, any> = {},
  UnsafeMetadata extends Record<string, any> = {},
> = {
  public_metadata?: PublicMetadata;
  unsafe_metadata?: UnsafeMetadata;
};

/**
 * A user's full profile as returned by the flow API.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {string} user_id - The user's unique ID.
 * @property {WebauthnCredential[]} [passkeys] - The user's registered passkeys.
 * @property {WebauthnCredential[]} [security_keys] - The user's registered security keys.
 * @property {MFAConfig} [mfa_config] - The user's multi-factor authentication configuration.
 * @property {Email[]} [emails] - The user's email addresses.
 * @property {Username} [username] - The user's username, if set.
 * @property {UserMetadata} [metadata] - The user's custom metadata.
 * @property {Identity[]} [identities] - Third-party identities linked to the user.
 * @property {string} created_at - Timestamp of when the user was created.
 * @property {string} updated_at - Timestamp of when the user was last updated.
 * @property {string} [name] - The user's full name, if known (e.g. from a third-party provider).
 * @property {string} [given_name] - The user's given name, if known.
 * @property {string} [family_name] - The user's family name, if known.
 * @property {string} [picture] - A URL to the user's profile picture, if known.
 */
export interface User {
  readonly user_id: string;
  readonly passkeys?: WebauthnCredential[];
  readonly security_keys?: WebauthnCredential[];
  readonly mfa_config?: MFAConfig;
  readonly emails?: Email[];
  readonly username?: Username;
  readonly metadata?: UserMetadata;
  readonly identities?: Identity[];
  readonly created_at: string;
  readonly updated_at: string;
  readonly name?: string;
  readonly given_name?: string;
  readonly family_name?: string;
  readonly picture?: string;
}

/**
 * One of a user's active sessions.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {string} id - The session's unique ID.
 * @property {string} [user_agent] - The parsed user agent of the client that created the session.
 * @property {string} [user_agent_raw] - The raw user agent string of the client that created the session.
 * @property {string} [ip_address] - The IP address the session was created from.
 * @property {string} created_at - Timestamp of when the session was created.
 * @property {string} last_used - Timestamp of when the session was last used.
 * @property {boolean} current - Whether this is the session the current request is authenticated with.
 */
export interface Session {
  readonly id: string;
  readonly user_agent?: string;
  readonly user_agent_raw?: string;
  readonly ip_address?: string;
  readonly created_at: string;
  readonly last_used: string;
  readonly current: boolean;
}

/**
 * Payload of the `profile_init` state, the entry point of the account/profile management flow.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {User} user - The current user's profile.
 * @property {Session[]} [sessions] - The user's active sessions.
 */
export interface ProfilePayload {
  readonly user: User;
  readonly sessions?: Session[];
}

/**
 * The method used to complete a login.
 * @category SDK
 * @subcategory Flow Payloads
 */
export type LoginMethod = "password" | "passkey" | "passcode" | "third_party";

/**
 * The second-factor method used to complete a login, if any.
 * @category SDK
 * @subcategory Flow Payloads
 */
export type MFAMethod = "totp" | "security_key";

/**
 * Describes how a user most recently logged in.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {LoginMethod} login_method - The primary method used to log in.
 * @property {MFAMethod} [mfa_method] - The second-factor method used, if any.
 * @property {string} [third_party_provider] - The third-party provider used, if `login_method` is `third_party`.
 */
export interface LastLogin {
  readonly login_method: LoginMethod;
  readonly mfa_method?: MFAMethod;
  readonly third_party_provider?: string;
}

/**
 * Payload of the terminal `success` state, marking the end of a flow.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {User} user - The authenticated user.
 * @property {LastLogin} [last_login] - Details of how the user logged in, if applicable.
 * @property {Claims} claims - The claims associated with the newly created session.
 */
export interface SuccessPayload {
  readonly user: User;
  readonly last_login?: LastLogin;
  readonly claims: Claims;
}

/**
 * Payload of the `thirdparty` state, carrying the URL to redirect the user to for the OAuth flow.
 * @interface
 * @category SDK
 * @subcategory Flow Payloads
 * @property {string} redirect_url - The URL to redirect the browser to in order to start the third-party OAuth flow.
 */
export interface ThirdPartyPayload {
  readonly redirect_url: string;
}
