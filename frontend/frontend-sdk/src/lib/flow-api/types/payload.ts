import {
  CredentialCreationOptionsJSON,
  CredentialRequestOptionsJSON,
} from "@github/webauthn-json/src/webauthn-json/basic/json";

export interface PasscodeConfirmationPayload {
  readonly passcode_resent: boolean;
  readonly resend_after: number;
}

export interface LoginPasskeyPayload {
  readonly request_options: CredentialRequestOptionsJSON;
}

export interface MFAOTPSecretCreationPayload {
  readonly otp_secret: string;
  readonly otp_image_source: string;
}

export interface OnboardingVerifyPasskeyAttestationPayload {
  readonly creation_options: CredentialCreationOptionsJSON;
}

export interface LoginInitPayload {
  readonly request_options?: CredentialRequestOptionsJSON;
}

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

export interface Username {
  id: string;
  username: string;
  created_at: string;
  updated_at: string;
}

export interface Identity {
  readonly id: string;
  readonly provider: string;
}

export interface Email {
  readonly id: string;
  readonly address: string;
  readonly is_verified: boolean;
  readonly is_primary: boolean;
  readonly identities?: Identity[];
}

export interface MFAConfig {
  readonly auth_app_set_up: boolean;
  readonly totp_enabled: boolean;
  readonly security_keys_enabled: boolean;
}

export interface User {
  readonly user_id: string;
  readonly passkeys?: WebauthnCredential[];
  readonly security_keys?: WebauthnCredential[];
  readonly mfa_config?: MFAConfig;
  readonly emails?: Email[];
  readonly username?: Username;
  readonly created_at: string;
  readonly updated_at: string;
}

export interface ProfilePayload {
  readonly user: User;
}

export type LoginMethod = "password" | "passkey" | "passcode" | "third_party";

export type MFAMethod = "totp" | "security_key";

export interface LastLogin {
  readonly login_method: LoginMethod;
  readonly mfa_method?: MFAMethod;
  readonly third_party_provider?: string;
}

export interface SuccessPayload {
  readonly user: User;
  readonly last_login: LastLogin;
}

export interface ThirdPartyPayload {
  readonly redirect_url: string;
}
