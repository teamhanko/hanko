import { FlowError } from "./flowError";
import {
  PublicKeyCredentialWithAssertionJSON,
  PublicKeyCredentialWithAttestationJSON,
} from "@github/webauthn-json";

/**
 * Describes a single input field expected by an {@link Action}, including its constraints and any
 * validation error from a previous submission attempt.
 * @template TValue - The type of the input's value.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {string} name - The name of the input field.
 * @property {string} type - The input's type (e.g. `string`, `boolean`), as reported by the backend.
 * @property {TValue} [value] - The current or previously submitted value, if any.
 * @property {number} [min_length] - The minimum allowed length, for string inputs.
 * @property {number} [max_length] - The maximum allowed length, for string inputs.
 * @property {boolean} [required] - Indicates whether the input must be provided to run the action.
 * @property {boolean} [hidden] - Indicates whether the input should not be shown in a UI.
 * @property {FlowError} [error] - A validation error for this specific input, if the last submission failed.
 * @property {AllowedInputValues[]} [allowed_values] - The set of allowed values, if the input is restricted to a fixed list.
 */
export interface Input<TValue> {
  readonly name: string;
  readonly type: string;
  readonly value?: TValue;
  readonly min_length?: number;
  readonly max_length?: number;
  readonly required?: boolean;
  readonly hidden?: boolean;
  readonly error?: FlowError;
  readonly allowed_values?: AllowedInputValues[];
}

/**
 * One entry of a restricted set of values an {@link Input} may be set to.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {string} name - A human-readable label for the value.
 * @property {string} value - The value to submit when this option is chosen.
 */
export interface AllowedInputValues {
  readonly name: string;
  readonly value: string;
}

/**
 * Inputs for submitting a new password during password recovery.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} new_password - The new password to set.
 */
export interface PasswordRecoveryInputs {
  readonly new_password: Input<string>;
}

/**
 * Inputs for submitting a WebAuthn assertion response (passkey/security key login).
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<PublicKeyCredentialWithAssertionJSON>} assertion_response - The assertion response produced by the browser's WebAuthn API.
 */
export interface WebauthnVerifyAssertionResponseInputs {
  readonly assertion_response: Input<PublicKeyCredentialWithAssertionJSON>;
}

/**
 * Inputs for submitting a WebAuthn attestation response (passkey/security key registration).
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<PublicKeyCredentialWithAttestationJSON>} public_key - The attestation response produced by the browser's WebAuthn API.
 */
export interface WebauthnVerifyAttestationResponseInputs {
  readonly public_key: Input<PublicKeyCredentialWithAttestationJSON>;
}

/**
 * Inputs for submitting a login identifier during registration.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} [email] - The email address to register with.
 * @property {Input<string>} [username] - The username to register with.
 */
export interface RegisterLoginIdentifierInputs {
  readonly email?: Input<string>;
  readonly username?: Input<string>;
}

/**
 * Inputs for submitting a new password during registration.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} new_password - The password to register with.
 */
export interface RegisterPasswordInputs {
  readonly new_password: Input<string>;
}

/**
 * Inputs for reporting the client's WebAuthn capabilities to the backend.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<boolean>} webauthn_available - Whether the WebAuthn API is available in the browser.
 * @property {Input<boolean>} webauthn_conditional_mediation_available - Whether conditional mediation (passkey autofill) is supported.
 * @property {Input<boolean>} webauthn_platform_authenticator_available - Whether a platform authenticator (e.g. Touch ID, Windows Hello) is available.
 */
export interface RegisterClientCapabilitiesInputs {
  readonly webauthn_available: Input<boolean>;
  readonly webauthn_conditional_mediation_available: Input<boolean>;
  readonly webauthn_platform_authenticator_available: Input<boolean>;
}

/**
 * Inputs for submitting a login identifier to continue a login.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} [identifier] - A generic identifier, when the backend does not distinguish between email and username.
 * @property {Input<string>} [email] - The email address to log in with.
 * @property {Input<string>} [username] - The username to log in with.
 */
export interface ContinueWithLoginIdentifierInputs {
  readonly identifier?: Input<string>;
  readonly email?: Input<string>;
  readonly username?: Input<string>;
}

/**
 * Inputs for submitting a one-time passcode for verification.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} code - The passcode the user received (e.g. via email).
 */
export interface VerifyPasscodeInputs {
  readonly code: Input<string>;
}

/**
 * Inputs for adding a new email address to the account.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} email - The email address to add.
 */
export interface EmailCreateInputs {
  readonly email: Input<string>;
}

/**
 * Inputs for removing an email address from the account.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} email_id - The ID of the email address to remove.
 */
export interface EmailDeleteInputs {
  readonly email_id: Input<string>;
}

/**
 * Inputs for marking an email address as the primary one.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} email_id - The ID of the email address to set as primary.
 */
export interface EmailSetPrimaryInputs {
  readonly email_id: Input<string>;
}

/**
 * Inputs for requesting verification of an email address.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} email_id - The ID of the email address to verify.
 */
export interface EmailVerifyInputs {
  readonly email_id: Input<string>;
}

/**
 * Inputs for submitting a password (e.g. for login).
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} password - The password to submit.
 */
export interface PasswordInputs {
  readonly password: Input<string>;
}

/**
 * Inputs for patching the user's public/unsafe metadata.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<Record<string, any>>} patch_metadata - The metadata patch to apply.
 */
export interface PatchMetadataInputs {
  readonly patch_metadata: Input<Record<string, any>>;
}

/**
 * Inputs for setting the account's username.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} username - The username to set.
 */
export interface UsernameSetInputs {
  readonly username: Input<string>;
}

/**
 * Inputs for removing a registered security key.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} security_key_id - The ID of the security key to remove.
 */
export interface SecurityKeyDeleteInputs {
  readonly security_key_id: Input<string>;
}

/**
 * Inputs for renaming a registered passkey.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} passkey_id - The ID of the passkey to rename.
 * @property {Input<string>} passkey_name - The new name for the passkey.
 */
export interface PasskeyCredentialRenameInputs {
  readonly passkey_id: Input<string>;
  readonly passkey_name: Input<string>;
}

/**
 * Inputs for removing a registered passkey.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} passkey_id - The ID of the passkey to remove.
 */
export interface PasskeyCredentialDeleteInputs {
  readonly passkey_id: Input<string>;
}

/**
 * Inputs for exchanging a third-party OAuth callback token for a session.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} token - The token received from the third-party provider's callback.
 * @property {Input<string>} [code_verifier] - The PKCE code verifier matching the original authorization request, if PKCE was used.
 */
export interface ExchangeTokenInputs {
  readonly token: Input<string>;
  readonly code_verifier?: Input<string>;
}

/**
 * Inputs for starting a third-party OAuth login/registration/account-linking flow.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} provider - The third-party provider to use.
 * @property {Input<string>} redirect_to - The URL to redirect back to after the OAuth flow completes.
 * @property {Input<string>} [code_verifier] - The PKCE code verifier to use for the authorization request, if PKCE is enabled.
 */
export interface ThirdpartyOauthInputs {
  readonly provider: Input<string>;
  readonly redirect_to: Input<string>;
  readonly code_verifier?: Input<string>;
}

/**
 * Inputs for unlinking a third-party OAuth provider from the account.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} identity_id - The ID of the linked identity to remove.
 */
export interface DisconnectThirdpartyInputs {
  readonly identity_id: Input<string>;
}

/**
 * Inputs for setting whether the current device should be remembered for future logins.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<boolean>} remember_me - Whether to remember the device.
 */
export interface RememberMeInputs {
  readonly remember_me: Input<boolean>;
}

/**
 * Inputs for revoking one of the user's active sessions.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} session_id - The ID of the session to revoke.
 */
export interface SessionDeleteInputs {
  readonly session_id: Input<string>;
}

/**
 * Inputs for submitting a one-time password (TOTP) code.
 * @interface
 * @category SDK
 * @subcategory Flow Inputs
 * @property {Input<string>} otp_code - The code generated by the user's authenticator app.
 */
export interface OTPCodeInputs {
  readonly otp_code: Input<string>;
}
