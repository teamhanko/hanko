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

export interface OnboardingVerifyPasskeyAttestationPayload {
  readonly creation_options: CredentialCreationOptionsJSON;
}

export interface LoginInitPayload {
  readonly request_options?: CredentialRequestOptionsJSON;
}

export interface Passkey {
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

export interface User {
  readonly user_id: string;
  readonly passkeys?: Passkey[];
  readonly emails?: Email[];
  readonly username?: Username;
  readonly created_at: string;
  readonly updated_at: string;
}

export interface ProfilePayload {
  readonly user: User;
}

export interface SuccessPayload {
  readonly user: User;
}

export interface ThirdPartyPayload {
  readonly redirect_url: string;
}
