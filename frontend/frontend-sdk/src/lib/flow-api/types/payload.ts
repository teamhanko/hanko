import {
  CredentialCreationOptionsJSON,
  CredentialRequestOptionsJSON
} from "@github/webauthn-json/src/webauthn-json/basic/json";

interface LoginPasskeyPayload {
  readonly request_options: CredentialRequestOptionsJSON;
}

interface OnboardingVerifyPasskeyAttestationPayload {
  creation_options: CredentialCreationOptionsJSON;
}

export type { LoginPasskeyPayload, OnboardingVerifyPasskeyAttestationPayload };
