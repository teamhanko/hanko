import { Error } from "./error";
import {
  CredentialCreationOptionsJSON,
  PublicKeyCredentialWithAssertionJSON
} from "@github/webauthn-json/src/webauthn-json/basic/json";

interface Input<TValue> {
  name: string;
  type: string;
  value?: TValue;
  min_length?: number;
  max_length?: number;
  required?: boolean;
  hidden?: boolean;
  error?: Error;
}

interface PasswordRecoveryInputs {
  readonly new_password: Input<string>;
}

interface WebauthnVerifyAssertionResponseInputs {
  readonly assertion_response: Input<PublicKeyCredentialWithAssertionJSON>;
}

interface WebauthnVerifyAttestationResponseInputs {
  readonly public_key: Input<CredentialCreationOptionsJSON>;
}

interface RegisterLoginIdentifierInputs {
  readonly email?: Input<string>;
  readonly username?: Input<string>;
}

interface RegisterPasswordInputs {
  readonly new_password: Input<string>;
}

interface RegisterClientCapabilitiesInputs {
  readonly webauthn_available: Input<boolean>;
}

interface ContinueWithLoginIdentifierInputs {
  readonly identifier: Input<string>;
}

interface PasswordLoginInputs {
  readonly password: Input<string>;
}

interface VerifyPasscodeInputs {
  readonly code: Input<string>;
}

export type {
  Input,
  PasswordRecoveryInputs,
  WebauthnVerifyAssertionResponseInputs,
  WebauthnVerifyAttestationResponseInputs,
  RegisterLoginIdentifierInputs,
  RegisterPasswordInputs,
  RegisterClientCapabilitiesInputs,
  ContinueWithLoginIdentifierInputs,
  PasswordLoginInputs,
  VerifyPasscodeInputs,
};
