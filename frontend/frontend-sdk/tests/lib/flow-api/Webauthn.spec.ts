import { autoSteps } from "../../../src/lib/flow-api/auto-steps";
import WebauthnManager from "../../../src/lib/flow-api/WebauthnManager";

// Regression tests for the passkey credential-creation auto-steps. Canceling the
// passkey prompt rejects with a NotAllowedError (and superseded requests reject
// with an AbortError); neither means the credential already exists, so they must
// not be surfaced as "webauthn_credential_already_exists".
describe("autoSteps webauthn credential creation error handling", () => {
  const alreadyExistsCode = "webauthn_credential_already_exists";

  const buildState = (overrides: Record<string, unknown> = {}) => ({
    payload: { creation_options: {} },
    error: undefined,
    actions: {
      back: { run: jest.fn().mockResolvedValue({ name: "back_state" }) },
      webauthn_verify_attestation_response: { run: jest.fn() },
    },
    ...overrides,
  });

  const mockCreate = (impl: () => Promise<unknown>) => {
    jest.spyOn(WebauthnManager, "getInstance").mockReturnValue({
      createWebauthnCredential: jest.fn(impl),
    } as never);
  };

  afterEach(() => {
    jest.restoreAllMocks();
  });

  it("does not report 'already exists' when the user cancels the prompt", async () => {
    mockCreate(() =>
      Promise.reject(new DOMException("The operation was cancelled.", "NotAllowedError")),
    );
    const state = buildState();

    const result = await autoSteps.webauthn_credential_verification(state as never);

    expect(state.actions.back.run).toHaveBeenCalled();
    expect(result.error).toBeUndefined();
  });

  it("does not report 'already exists' when the request is aborted", async () => {
    mockCreate(() =>
      Promise.reject(new DOMException("The operation was aborted.", "AbortError")),
    );
    const state = buildState();

    const result = await autoSteps.webauthn_credential_verification(state as never);

    expect(result.error).toBeUndefined();
  });

  it("still reports 'already exists' for an InvalidStateError", async () => {
    mockCreate(() =>
      Promise.reject(new DOMException("Credential already registered.", "InvalidStateError")),
    );
    const state = buildState();

    const result = await autoSteps.webauthn_credential_verification(state as never);

    expect(result.error).toEqual({
      code: alreadyExistsCode,
      message: "Webauthn credential already exists",
    });
  });

  it("verifies the attestation response on success", async () => {
    const attestation = { id: "cred" };
    mockCreate(() => Promise.resolve(attestation));
    const state = buildState();

    await autoSteps.webauthn_credential_verification(state as never);

    expect(
      state.actions.webauthn_verify_attestation_response.run,
    ).toHaveBeenCalledWith({ public_key: attestation });
    expect(state.actions.back.run).not.toHaveBeenCalled();
  });
});
