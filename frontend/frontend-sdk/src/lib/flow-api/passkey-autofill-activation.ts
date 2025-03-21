import { PasskeyAutofillActivationHandlers } from "./types/flow";
import WebauthnManager from "./WebauthnManager";

export const passkeyAutofillActivationHandlers: PasskeyAutofillActivationHandlers =
  {
    login_init: async (state) => {
      return void (async function () {
        const manager = WebauthnManager.getInstance();

        if (state.payload.request_options) {
          try {
            const { publicKey } = state.payload.request_options;

            const assertionResponse =
              await manager.getConditionalWebauthnCredential(publicKey);

            return await state.actions.webauthn_verify_assertion_response.run({
              assertion_response: assertionResponse,
            });
          } catch {
            // We do not need to handle the error, because this is a conditional request, which can fail silently
            return;
          }
        }
      })();
    },
  };
