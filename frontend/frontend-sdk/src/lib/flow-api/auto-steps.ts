import { AutoSteps } from "./types/flow";
import { WebauthnSupport } from "../WebauthnSupport";
import { PublicKeyCredentialWithAssertionJSON } from "@github/webauthn-json/src/webauthn-json/basic/json";
import { get as getWebauthnCredential } from "@github/webauthn-json/dist/types/basic/api";

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
    } catch (error) {
      return await state.actions.back.run(null);
    }

    return await state.actions.webauthn_verify_assertion_response.run({
      assertion_response: assertionResponse,
    });
  },
};
