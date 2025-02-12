import { AutoSteps } from "./types/flow";
import { WebauthnSupport } from "../WebauthnSupport";
import WebauthnManager from "./WebauthnManager";

// Helper function to handle WebAuthn credential creation and error handling
// eslint-disable-next-line require-jsdoc
async function handleCredentialCreation(
  state: any,
  manager: WebauthnManager,
  options: any,
  errorCode: string = "webauthn_credential_already_exists",
  errorMessage: string = "Webauthn credential already exists",
) {
  try {
    const attestationResponse = await manager.createWebauthnCredential(options);
    return await state.actions.webauthn_verify_attestation_response.run({
      public_key: attestationResponse,
    });
  } catch {
    const nextState = await state.actions.back.run();
    nextState.error = { code: errorCode, message: errorMessage };
    return nextState;
  }
}

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
    const manager = WebauthnManager.getInstance();
    try {
      const assertionResponse = await manager.getWebauthnCredential(
        state.payload.request_options,
      );
      return await state.actions.webauthn_verify_assertion_response.run({
        assertion_response: assertionResponse,
      });
    } catch (e) {
      const nextState = await state.actions.back.run();
      if (state.error) {
        nextState.error = state.error;
      }
      return nextState;
    }
  },

  onboarding_verify_passkey_attestation: async (state) => {
    const manager = WebauthnManager.getInstance();
    return handleCredentialCreation(
      state,
      manager,
      state.payload.creation_options,
    );
  },

  webauthn_credential_verification: async (state) => {
    const manager = WebauthnManager.getInstance();
    return handleCredentialCreation(
      state,
      manager,
      state.payload.creation_options,
    );
  },

  thirdparty: async (state) => {
    const searchParams = new URLSearchParams(window.location.search);
    const token = searchParams.get("hanko_token");

    if (token && token.length > 0) {
      const nextState = await state.actions.exchange_token.run(
        { token },
        { dispatchAfterStateChangeEvent: false },
      );

      searchParams.delete("hanko_token");
      history.replaceState(
        null,
        null,
        window.location.pathname + searchParams.toString(),
      );

      nextState.dispatchAfterStateChangeEvent();
      return nextState;
    }

    state.saveToLocalStorage();
    window.location.assign(state.payload.redirect_url);
    return state;
  },

  success: async (state) => {
    state.hanko.relay.dispatchSessionCreatedEvent(state.hanko.session.get());
    return Promise.resolve(state);
  },

  account_deleted: async (state) => {
    state.hanko.relay.dispatchUserDeletedEvent();
    return Promise.resolve(state);
  },
};
