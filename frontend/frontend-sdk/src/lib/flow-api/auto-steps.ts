import { AutoSteps, DefaultHandlers } from "./types/flow";
import { WebauthnSupport } from "../WebauthnSupport";
import {
  PublicKeyCredentialWithAssertionJSON,
  PublicKeyCredentialWithAttestationJSON,
} from "@github/webauthn-json/src/webauthn-json/basic/json";
import {
  create as createWebauthnCredential,
  get as getWebauthnCredential,
} from "@github/webauthn-json/dist/types/basic/api";

let webauthnAbortController = new AbortController();

const createWebauthnAbortSignal = () => {
  if (webauthnAbortController) {
    webauthnAbortController.abort();
  }

  webauthnAbortController = new AbortController();
  return webauthnAbortController.signal;
};

export const autoSteps: AutoSteps = {
  preflight: async (state) => {
    return await state.actions.register_client_capabilities.run({
      webauthn_available: WebauthnSupport.supported(),
      webauthn_conditional_mediation_available:
        await WebauthnSupport.isConditionalMediationAvailable(),
      webauthn_platform_authenticator_available:
        await WebauthnSupport.isPlatformAuthenticatorAvailable(),
    });
  },
  login_passkey: async (state) => {
    let assertionResponse: PublicKeyCredentialWithAssertionJSON;

    try {
      assertionResponse = await getWebauthnCredential({
        ...state.payload.request_options,
        signal: createWebauthnAbortSignal(),
      });
    } catch {
      const nextState = await state.actions.back.run(null);

      if (state.error) {
        nextState.error = state.error;
      }

      return nextState;
    }

    return await state.actions.webauthn_verify_assertion_response.run({
      assertion_response: assertionResponse,
    });
  },
  onboarding_verify_passkey_attestation: async (state) => {
    let attestationResponse: PublicKeyCredentialWithAttestationJSON;

    try {
      attestationResponse = await createWebauthnCredential({
        ...state.payload.creation_options,
        signal: createWebauthnAbortSignal(),
      });
    } catch {
      const nextState = await state.actions.back.run(null);

      nextState.error = {
        code: "webauthn_credential_already_exists",
        message: "Webauthn credential already exists",
      };

      return nextState;
    }

    return await state.actions.webauthn_verify_attestation_response.run({
      public_key: attestationResponse,
    });
  },
  webauthn_credential_verification: async (state) => {
    let attestationResponse: PublicKeyCredentialWithAttestationJSON;

    try {
      attestationResponse = await createWebauthnCredential({
        ...state.payload.creation_options,
        signal: createWebauthnAbortSignal(),
      });
    } catch {
      const nextState = await state.actions.back.run(null);

      nextState.error = {
        code: "webauthn_credential_already_exists",
        message: "Webauthn credential already exists",
      };

      return nextState;
    }

    return await state.actions.webauthn_verify_attestation_response.run({
      public_key: attestationResponse,
    });
  },
};

export const defaultHandlers: DefaultHandlers = {
  login_init: async (state) => {
    return void (async function () {
      if (state.payload.request_options) {
        let assertionResponse: PublicKeyCredentialWithAssertionJSON;

        try {
          assertionResponse = await getWebauthnCredential({
            publicKey: state.payload.request_options.publicKey,
            mediation: "conditional" as CredentialMediationRequirement,
            signal: createWebauthnAbortSignal(),
          });
        } catch (error) {
          // We do not need to handle the error, because this is a conditional request, which can fail silently
          return;
        }

        return await state.actions.webauthn_verify_assertion_response.run({
          assertion_response: assertionResponse,
        });
      }
    })();
  },
  success: async (state) => {
    // if (state.payload?.last_login) {
    //   localStorage.setItem(
    //     storageKeyLastLogin,
    //     JSON.stringify(state.payload.last_login),
    //   );
    // }
    // hanko.relay.dispatchSessionCreatedEvent(hanko.session.get());
    return;
  },
};
