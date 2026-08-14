import {
  ContinueWithLoginIdentifierInputs,
  EmailCreateInputs,
  EmailDeleteInputs,
  EmailSetPrimaryInputs,
  EmailVerifyInputs,
  ExchangeTokenInputs,
  PasskeyCredentialDeleteInputs,
  PasskeyCredentialRenameInputs,
  PasswordRecoveryInputs,
  PasswordInputs,
  PatchMetadataInputs,
  RegisterClientCapabilitiesInputs,
  RegisterLoginIdentifierInputs,
  RegisterPasswordInputs,
  ThirdpartyOauthInputs,
  UsernameSetInputs,
  VerifyPasscodeInputs,
  WebauthnVerifyAssertionResponseInputs,
  WebauthnVerifyAttestationResponseInputs,
  SessionDeleteInputs,
  OTPCodeInputs,
  SecurityKeyDeleteInputs,
  RememberMeInputs,
  DisconnectThirdpartyInputs,
} from "./input";

/**
 * Represents a single action that can be run from a {@link State}. `action` and `href` identify
 * the action for the backend, `inputs` describes the input fields the action accepts (see
 * {@link State#actions} for how these map to a concrete state), and `description` is a
 * human-readable summary of what running the action does.
 * @template TInputs - The shape of the inputs this action accepts, or `null` if it accepts none.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {string} action - The unique name of the action within the current flow.
 * @property {string} href - The URL the action is submitted to when run.
 * @property {TInputs} inputs - The input fields accepted by the action.
 * @property {string} description - A human-readable description of the action.
 */
export interface Action<TInputs> {
  action: string;
  href: string;
  inputs: TInputs;
  description: string;
}

/**
 * Actions available in the `preflight` state, run once at the start of a flow to register client capabilities.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<RegisterClientCapabilitiesInputs>} register_client_capabilities - Reports client capabilities (e.g. WebAuthn support) to the backend.
 */
export interface PreflightActions {
  readonly register_client_capabilities: Action<RegisterClientCapabilitiesInputs>;
}

/**
 * Actions available in the `login_init` state, the entry point of a login flow.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<ContinueWithLoginIdentifierInputs>} [continue_with_login_identifier] - Submits an identifier (e.g. email or username) to continue the login.
 * @property {Action<null>} [webauthn_generate_request_options] - Requests WebAuthn assertion options to attempt a passkey login.
 * @property {Action<WebauthnVerifyAssertionResponseInputs>} [webauthn_verify_assertion_response] - Submits the WebAuthn assertion response for verification.
 * @property {Action<ThirdpartyOauthInputs>} [thirdparty_oauth] - Starts a third-party OAuth login.
 * @property {Action<RememberMeInputs>} [remember_me] - Sets whether the device should be remembered for future logins.
 */
export interface LoginInitActions {
  readonly continue_with_login_identifier?: Action<ContinueWithLoginIdentifierInputs>;
  readonly webauthn_generate_request_options?: Action<null>;
  readonly webauthn_verify_assertion_response?: Action<WebauthnVerifyAssertionResponseInputs>;
  readonly thirdparty_oauth?: Action<ThirdpartyOauthInputs>;
  readonly remember_me?: Action<RememberMeInputs>;
}

/**
 * Actions available in the `profile_init` state, the user's account/profile management screen.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<null>} [account_delete] - Deletes the current user's account.
 * @property {Action<null>} [continue_to_otp_secret_creation] - Continues to setting up an authenticator app (TOTP) for MFA.
 * @property {Action<EmailCreateInputs>} [email_create] - Adds a new email address to the account.
 * @property {Action<EmailDeleteInputs>} [email_delete] - Removes an email address from the account.
 * @property {Action<EmailVerifyInputs>} [email_verify] - Requests verification of an unverified email address.
 * @property {Action<EmailSetPrimaryInputs>} [email_set_primary] - Sets an email address as the primary one.
 * @property {Action<null>} [otp_secret_delete] - Removes the configured authenticator app (TOTP) secret.
 * @property {Action<PasswordInputs>} [password_create] - Sets a password for the account.
 * @property {Action<PasswordInputs>} [password_update] - Updates the account's existing password.
 * @property {Action<null>} [password_delete] - Removes the account's password.
 * @property {Action<PatchMetadataInputs>} patch_metadata - Updates the user's public/unsafe metadata.
 * @property {Action<null>} [security_key_create] - Starts registering a new hardware security key.
 * @property {Action<SecurityKeyDeleteInputs>} [security_key_delete] - Removes a registered security key.
 * @property {Action<UsernameSetInputs>} [username_create] - Sets a username for the account.
 * @property {Action<null>} [username_delete] - Removes the account's username.
 * @property {Action<UsernameSetInputs>} [username_update] - Updates the account's existing username.
 * @property {Action<null>} [webauthn_credential_create] - Starts registering a new passkey.
 * @property {Action<PasskeyCredentialRenameInputs>} [webauthn_credential_rename] - Renames a registered passkey.
 * @property {Action<PasskeyCredentialDeleteInputs>} [webauthn_credential_delete] - Removes a registered passkey.
 * @property {Action<WebauthnVerifyAttestationResponseInputs>} [webauthn_verify_attestation_response] - Submits the WebAuthn attestation response to complete a passkey/security key registration.
 * @property {Action<SessionDeleteInputs>} [session_delete] - Revokes one of the user's active sessions.
 * @property {Action<ThirdpartyOauthInputs>} [connect_thirdparty_oauth_provider] - Links a third-party OAuth provider to the account.
 * @property {Action<DisconnectThirdpartyInputs>} [disconnect_thirdparty_oauth_provider] - Unlinks a third-party OAuth provider from the account.
 */
export interface ProfileInitActions {
  readonly account_delete?: Action<null>;
  readonly continue_to_otp_secret_creation?: Action<null>;
  readonly email_create?: Action<EmailCreateInputs>;
  readonly email_delete?: Action<EmailDeleteInputs>;
  readonly email_verify?: Action<EmailVerifyInputs>;
  readonly email_set_primary?: Action<EmailSetPrimaryInputs>;
  readonly otp_secret_delete?: Action<null>;
  readonly password_create?: Action<PasswordInputs>;
  readonly password_update?: Action<PasswordInputs>;
  readonly password_delete?: Action<null>;
  readonly patch_metadata: Action<PatchMetadataInputs>;
  readonly security_key_create?: Action<null>;
  readonly security_key_delete?: Action<SecurityKeyDeleteInputs>;
  readonly username_create?: Action<UsernameSetInputs>;
  readonly username_delete?: Action<null>;
  readonly username_update?: Action<UsernameSetInputs>;
  readonly webauthn_credential_create?: Action<null>;
  readonly webauthn_credential_rename?: Action<PasskeyCredentialRenameInputs>;
  readonly webauthn_credential_delete?: Action<PasskeyCredentialDeleteInputs>;
  readonly webauthn_verify_attestation_response?: Action<WebauthnVerifyAttestationResponseInputs>;
  readonly session_delete?: Action<SessionDeleteInputs>;
  readonly connect_thirdparty_oauth_provider?: Action<ThirdpartyOauthInputs>;
  readonly disconnect_thirdparty_oauth_provider?: Action<DisconnectThirdpartyInputs>;
}

/**
 * Actions available in the `login_method_chooser` state, letting the user pick a login method.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<null>} [continue_to_password_login] - Switches to password-based login.
 * @property {Action<null>} [continue_to_passcode_confirmation] - Switches to passcode (one-time code) based login.
 * @property {Action<null>} [webauthn_generate_request_options] - Requests WebAuthn assertion options to attempt a passkey login.
 * @property {Action<null>} back - Returns to the previous state.
 */
export interface LoginMethodChooserActions {
  readonly continue_to_password_login?: Action<null>;
  readonly continue_to_passcode_confirmation?: Action<null>;
  readonly webauthn_generate_request_options?: Action<null>;
  readonly back: Action<null>;
}

/**
 * Actions available in the `login_otp` state, where the user submits a one-time passcode.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<OTPCodeInputs>} otp_code_validate - Submits the entered one-time code for validation.
 * @property {Action<null>} [continue_to_login_security_key] - Switches to security-key based login.
 */
export interface LoginOTPActions {
  readonly otp_code_validate: Action<OTPCodeInputs>;
  readonly continue_to_login_security_key?: Action<null>;
}

/**
 * Actions available in the `login_password` state.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<PasswordInputs>} password_login - Submits the entered password for login.
 * @property {Action<null>} [continue_to_passcode_confirmation_recovery] - Starts password recovery via a passcode.
 * @property {Action<null>} continue_to_login_method_chooser - Returns to the login method chooser.
 * @property {Action<null>} back - Returns to the previous state.
 */
export interface LoginPasswordActions {
  readonly password_login: Action<PasswordInputs>;
  readonly continue_to_passcode_confirmation_recovery?: Action<null>;
  readonly continue_to_login_method_chooser: Action<null>;
  readonly back: Action<null>;
}

/**
 * Actions available in the `login_password_recovery` state.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<PasswordRecoveryInputs>} password_recovery - Submits a new password to complete password recovery.
 */
export interface LoginPasswordRecoveryActions {
  readonly password_recovery: Action<PasswordRecoveryInputs>;
}

/**
 * Actions available in the `login_passkey` state, while a passkey login ceremony is in progress.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<WebauthnVerifyAssertionResponseInputs>} webauthn_verify_assertion_response - Submits the WebAuthn assertion response for verification.
 * @property {Action<null>} back - Returns to the previous state.
 */
export interface LoginPasskeyActions {
  readonly webauthn_verify_assertion_response: Action<WebauthnVerifyAssertionResponseInputs>;
  readonly back: Action<null>;
}

/**
 * Actions available in the `login_security_key` state, while a security-key login ceremony is in progress.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<null>} webauthn_generate_request_options - Requests WebAuthn assertion options for the security key.
 * @property {Action<null>} [continue_to_login_otp] - Switches to one-time-passcode based login.
 */
export interface LoginSecurityKeyActions {
  readonly webauthn_generate_request_options: Action<null>;
  readonly continue_to_login_otp?: Action<null>;
}

/**
 * Actions available in the `mfa_method_chooser` state, letting the user pick a second factor method.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<null>} [continue_to_otp_secret_creation] - Switches to authenticator app (TOTP) setup.
 * @property {Action<null>} [continue_to_security_key_creation] - Switches to security key setup.
 * @property {Action<null>} [skip] - Skips second factor setup, if optional.
 * @property {Action<null>} [back] - Returns to the previous state.
 */
export interface MFAMethodChooserActions {
  readonly continue_to_otp_secret_creation?: Action<null>;
  readonly continue_to_security_key_creation?: Action<null>;
  readonly skip?: Action<null>;
  readonly back?: Action<null>;
}

/**
 * Actions available in the `mfa_otp_secret_creation` state, while setting up an authenticator app (TOTP).
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<OTPCodeInputs>} otp_code_verify - Submits a code generated from the authenticator app to confirm setup.
 * @property {Action<null>} back - Returns to the previous state.
 */
export interface MFAAOTPSecretCreationActions {
  readonly otp_code_verify: Action<OTPCodeInputs>;
  readonly back: Action<null>;
}

/**
 * Actions available in the `mfa_security_key_creation` state, while registering a security key as a second factor.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<null>} webauthn_generate_creation_options - Requests WebAuthn attestation (registration) options for the security key.
 * @property {Action<null>} back - Returns to the previous state.
 */
export interface MFASecurityKeyCreationActions {
  readonly webauthn_generate_creation_options: Action<null>;
  readonly back: Action<null>;
}

/**
 * Actions available in the `onboarding_create_passkey` state, while registering a passkey during onboarding.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<null>} webauthn_generate_creation_options - Requests WebAuthn attestation (registration) options for the passkey.
 * @property {Action<null>} [skip] - Skips passkey creation, if optional.
 * @property {Action<null>} [back] - Returns to the previous state.
 */
export interface OnboardingCreatePasskeyActions {
  readonly webauthn_generate_creation_options: Action<null>;
  readonly skip?: Action<null>;
  readonly back?: Action<null>;
}

/**
 * Actions available in the `onboarding_verify_passkey_attestation` / `webauthn_credential_verification` states,
 * while a passkey or security key attestation response is being submitted.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<WebauthnVerifyAttestationResponseInputs>} webauthn_verify_attestation_response - Submits the WebAuthn attestation response for verification.
 * @property {Action<null>} back - Returns to the previous state.
 */
export interface OnboardingVerifyPasskeyAttestationActions {
  readonly webauthn_verify_attestation_response: Action<WebauthnVerifyAttestationResponseInputs>;
  readonly back: Action<null>;
}

/**
 * Actions available in the `registration_init` state, the entry point of a registration flow.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<RegisterLoginIdentifierInputs>} register_login_identifier - Submits an identifier (e.g. email or username) to start registration.
 * @property {Action<ThirdpartyOauthInputs>} [thirdparty_oauth] - Starts registration via a third-party OAuth provider.
 * @property {Action<RememberMeInputs>} [remember_me] - Sets whether the device should be remembered for future logins.
 */
export interface RegistrationInitActions {
  readonly register_login_identifier: Action<RegisterLoginIdentifierInputs>;
  readonly thirdparty_oauth?: Action<ThirdpartyOauthInputs>;
  readonly remember_me?: Action<RememberMeInputs>;
}

/**
 * Actions available in the `password_creation` state, while setting a password during registration.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<RegisterPasswordInputs>} register_password - Submits the chosen password.
 * @property {Action<null>} [back] - Returns to the previous state.
 * @property {Action<null>} [skip] - Skips setting a password, if optional.
 */
export interface PasswordCreationActions {
  readonly register_password: Action<RegisterPasswordInputs>;
  readonly back?: Action<null>;
  readonly skip?: Action<null>;
}

/**
 * Actions available in the `passcode_confirmation` state, while confirming a one-time passcode sent to the user.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<VerifyPasscodeInputs>} verify_passcode - Submits the entered passcode for verification.
 * @property {Action<null>} resend_passcode - Requests the passcode to be sent again.
 * @property {Action<null>} back - Returns to the previous state.
 */
export interface PasscodeConfirmationActions {
  readonly verify_passcode: Action<VerifyPasscodeInputs>;
  readonly resend_passcode: Action<null>;
  readonly back: Action<null>;
}

/**
 * Actions available in the `onboarding_email` state, while setting an email address during onboarding.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<EmailCreateInputs>} email_address_set - Submits the chosen email address.
 * @property {Action<null>} skip - Skips setting an email address, if optional.
 */
export interface OnboardingEmailActions {
  readonly email_address_set: Action<EmailCreateInputs>;
  readonly skip: Action<null>;
}

/**
 * Actions available in the `onboarding_username` state, while setting a username during onboarding.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<UsernameSetInputs>} username_create - Submits the chosen username.
 * @property {Action<null>} skip - Skips setting a username, if optional.
 */
export interface OnboardingUsernameActions {
  readonly username_create: Action<UsernameSetInputs>;
  readonly skip: Action<null>;
}

/**
 * Actions available in the `credential_onboarding_chooser` state, letting the user pick how to onboard a credential.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<null>} continue_to_passkey_registration - Switches to passkey registration.
 * @property {Action<null>} continue_to_password_registration - Switches to password registration.
 * @property {Action<null>} skip - Skips credential onboarding, if optional.
 * @property {Action<null>} back - Returns to the previous state.
 */
export interface CredentialOnboardingChooserActions {
  readonly continue_to_passkey_registration: Action<null>;
  readonly continue_to_password_registration: Action<null>;
  readonly skip: Action<null>;
  readonly back: Action<null>;
}

/**
 * Actions available in the `device_trust` state, letting the user decide whether to trust the current device.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<null>} trust_device - Marks the current device as trusted.
 * @property {Action<null>} skip - Declines to trust the current device.
 * @property {Action<null>} [back] - Returns to the previous state.
 */
export interface DeviceTrustActions {
  readonly trust_device: Action<null>;
  readonly skip: Action<null>;
  readonly back?: Action<null>;
}

/**
 * Actions available in the `thirdparty` state, while a third-party OAuth flow is being completed.
 * @interface
 * @category SDK
 * @subcategory Flow Actions
 * @property {Action<ExchangeTokenInputs>} exchange_token - Exchanges the third-party provider's callback token for a session.
 * @property {Action<null>} [back] - Returns to the previous state.
 */
export interface ThirdPartyActions {
  readonly exchange_token: Action<ExchangeTokenInputs>;
  readonly back?: Action<null>;
}
