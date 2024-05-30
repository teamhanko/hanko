import { Error } from "./error";
import {
  PublicKeyCredentialWithAttestationJSON,
  PublicKeyCredentialWithAssertionJSON,
} from "@github/webauthn-json";

export interface Input<TValue> {
  readonly name: string;
  readonly type: string;
  value?: TValue;
  readonly min_length?: number;
  readonly max_length?: number;
  readonly required?: boolean;
  readonly hidden?: boolean;
  readonly error?: Error;
}

export interface PasswordRecoveryInputs {
  readonly new_password: Input<string>;
}

export interface WebauthnVerifyAssertionResponseInputs {
  readonly assertion_response: Input<PublicKeyCredentialWithAssertionJSON>;
}

export interface WebauthnVerifyAttestationResponseInputs {
  readonly public_key: Input<PublicKeyCredentialWithAttestationJSON>;
}

export interface RegisterLoginIdentifierInputs {
  readonly email?: Input<string>;
  readonly username?: Input<string>;
}

export interface RegisterPasswordInputs {
  readonly new_password: Input<string>;
}

export interface RegisterClientCapabilitiesInputs {
  readonly webauthn_available: Input<boolean>;
  readonly webauthn_conditional_mediation_available: Input<boolean>;
}

export interface ContinueWithLoginIdentifierInputs {
  readonly identifier?: Input<string>;
  readonly email?: Input<string>;
  readonly username?: Input<string>;
}

export interface PasswordLoginInputs {
  readonly password: Input<string>;
}

export interface VerifyPasscodeInputs {
  readonly code: Input<string>;
}

export interface EmailCreateInputs {
  readonly email: Input<string>;
}

export interface EmailDeleteInputs {
  readonly email_id: Input<string>;
}

export interface EmailSetPrimaryInputs {
  readonly email_id: Input<string>;
}

export interface EmailVerifyInputs {
  readonly email_id: Input<string>;
}

export interface PasswordSetInputs {
  readonly password: Input<string>;
}

export interface UsernameSetInputs {
  readonly username: Input<string>;
}

export interface PasskeyCredentialRename {
  readonly passkey_id: Input<string>;
  readonly passkey_name: Input<string>;
}

export interface PasskeyCredentialDelete {
  readonly passkey_id: Input<string>;
}
