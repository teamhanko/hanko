import { AutoSteps } from "./types/flow";
import { WebauthnSupport } from "../WebauthnSupport";
import WebauthnManager from "./WebauthnManager";
import { CredentialCreationOptionsJSON } from "@github/webauthn-json";

// Helper function to handle WebAuthn credential creation and error handling
// eslint-disable-next-line require-jsdoc
async function handleCredentialCreation(
  state: any,
  manager: WebauthnManager,
  options: CredentialCreationOptionsJSON,
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
    } catch {
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

  async thirdparty(state) {
    const searchParams = new URLSearchParams(window.location.search);
    const token = searchParams.get("hanko_token");
    const error = searchParams.get("error");

    const updateUrl = (paramsToDelete: string[]) => {
      paramsToDelete.forEach((param) => searchParams.delete(param));
      const newSearch = searchParams.toString()
        ? `?${searchParams.toString()}`
        : "";
      history.replaceState(
        null,
        null,
        `${window.location.pathname}${newSearch}`,
      );
    };

    if (token?.length > 0) {
      updateUrl(["hanko_token"]);
      return await state.actions.exchange_token.run({ token });
    }

    if (error?.length > 0) {
      const errorCode =
        error === "access_denied"
          ? "third_party_access_denied"
          : "technical_error";
      const message = searchParams.get("error_description");

      updateUrl(["error", "error_description"]);

      const nextState = await state.actions.back.run(null, {
        dispatchAfterStateChangeEvent: false,
      });

      nextState.error = { code: errorCode, message };
      nextState.dispatchAfterStateChangeEvent();

      return nextState;
    }

    if (!state.readFromLocalStorage) {
      state.saveToLocalStorage();
      window.location.assign(state.payload.redirect_url);
    } else {
      return await state.actions.back.run();
    }

    return state;
  },

  success: async (state) => {
    const { claims } = state.payload;
    const expirationSeconds = Date.parse(claims.expiration) - Date.now();
    state.removeFromLocalStorage();
    state.hanko.relay.dispatchSessionCreatedEvent({
      claims,
      expirationSeconds,
    });
    return state;
  },

  account_deleted: async (state) => {
    state.removeFromLocalStorage();
    state.hanko.relay.dispatchUserDeletedEvent();
    return state;
  },
};
