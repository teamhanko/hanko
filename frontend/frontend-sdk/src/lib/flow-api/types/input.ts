import { Error } from "./error";
import {
  CredentialCreationOptionsJSON,
  PublicKeyCredentialWithAssertionJSON,
} from "@github/webauthn-json/src/webauthn-json/basic/json";

export interface Input<TValue> {
  name: string;
  type: string;
  value?: TValue;
  min_length?: number;
  max_length?: number;
  required?: boolean;
  hidden?: boolean;
  error?: Error;
}

export interface PasswordRecoveryInputs {
  readonly new_password: Input<string>;
}

export interface WebauthnVerifyAssertionResponseInputs {
  readonly assertion_response: Input<PublicKeyCredentialWithAssertionJSON>;
}

export interface WebauthnVerifyAttestationResponseInputs {
  readonly public_key: Input<CredentialCreationOptionsJSON>;
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
}

export interface ContinueWithLoginIdentifierInputs {
  readonly identifier: Input<string>;
}

export interface PasswordLoginInputs {
  readonly password: Input<string>;
}

export interface VerifyPasscodeInputs {
  readonly code: Input<string>;
}
