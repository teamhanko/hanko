import { CredentialCreationOptionsJSON } from "@github/webauthn-json";
import WebauthnManager from "../../../src/lib/flow-api/WebauthnManager";
import { RequestTimeoutError } from "../../../src/lib/Errors";

// Well beyond any deadline the ceremony can be configured with, so the
// assertions are about the ceremony settling at all, not about a specific value.
const PROBE_MS = 30 * 60 * 1000;
const SERVER_TIMEOUT_MS = 60 * 1000;

type Outcome =
  | { status: "pending" }
  | { status: "resolved" }
  | { status: "rejected"; error: unknown };

const creationOptions = (timeout?: number): CredentialCreationOptionsJSON => ({
  publicKey: {
    rp: { name: "Hanko" },
    user: { id: "dXNlcg", name: "user@example.com", displayName: "user" },
    challenge: "Y2hhbGxlbmdl",
    pubKeyCredParams: [{ type: "public-key", alg: -7 }],
    timeout,
  },
});

const createdCredential = () => ({
  type: "public-key",
  id: "credential-id",
  rawId: new Uint8Array([1, 2, 3]).buffer,
  authenticatorAttachment: "platform",
  response: {
    clientDataJSON: new Uint8Array([4, 5, 6]).buffer,
    attestationObject: new Uint8Array([7, 8, 9]).buffer,
  },
  getClientExtensionResults: () => ({}),
});

// Some platform authenticators leave `navigator.credentials.create()` pending
// forever and do not honor the WebAuthn `timeout` themselves. Without a
// client-side deadline the awaiting caller never resumes, which leaves the
// passkey step disabled with no way to retry, skip or leave it.
describe("WebauthnManager credential creation deadline", () => {
  let credentials: { create: jest.Mock; get: jest.Mock };

  const runCreation = (options: CredentialCreationOptionsJSON) => {
    const outcome: { current: Outcome } = { current: { status: "pending" } };
    void WebauthnManager.getInstance()
      .createWebauthnCredential(options)
      .then(
        () => (outcome.current = { status: "resolved" }),
        (error) => (outcome.current = { status: "rejected", error }),
      );
    return outcome;
  };

  beforeEach(() => {
    credentials = navigator.credentials as never;
    credentials.create.mockReset();
  });

  it("rejects with a timeout error when the ceremony never settles", async () => {
    credentials.create.mockImplementation(() => new Promise(() => {}));

    const outcome = runCreation(creationOptions());
    await jest.advanceTimersByTimeAsync(PROBE_MS);

    expect(outcome.current).toEqual({
      status: "rejected",
      error: expect.any(RequestTimeoutError),
    });
  });

  it("aborts the pending ceremony at the timeout given in the creation options", async () => {
    let signal: AbortSignal;
    credentials.create.mockImplementation(
      (options: CredentialCreationOptions) => {
        signal = options.signal;
        return new Promise(() => {});
      },
    );

    const outcome = runCreation(creationOptions(SERVER_TIMEOUT_MS));
    await jest.advanceTimersByTimeAsync(SERVER_TIMEOUT_MS);

    expect(signal.aborted).toBe(true);
    expect(outcome.current.status).toBe("rejected");
  });

  it("treats a zero timeout as unset instead of ending the ceremony at once", async () => {
    credentials.create.mockImplementation(() => new Promise(() => {}));

    const outcome = runCreation(creationOptions(0));
    await jest.advanceTimersByTimeAsync(SERVER_TIMEOUT_MS);

    expect(outcome.current.status).toBe("pending");
  });

  it("aborts only its own ceremony when a later one has superseded it", async () => {
    const signals: AbortSignal[] = [];
    credentials.create.mockImplementation(
      (options: CredentialCreationOptions) => {
        signals.push(options.signal);
        return new Promise(() => {});
      },
    );

    const superseded = runCreation(creationOptions(SERVER_TIMEOUT_MS));
    const current = runCreation(creationOptions(PROBE_MS));
    await jest.advanceTimersByTimeAsync(SERVER_TIMEOUT_MS);

    expect(superseded.current.status).toBe("rejected");
    expect(current.current.status).toBe("pending");
    expect(signals[1].aborted).toBe(false);
  });

  it("does not abort a ceremony that completed before the deadline", async () => {
    let signal: AbortSignal;
    credentials.create.mockImplementation(
      (options: CredentialCreationOptions) => {
        signal = options.signal;
        return Promise.resolve(createdCredential());
      },
    );

    await WebauthnManager.getInstance().createWebauthnCredential(
      creationOptions(),
    );
    await jest.advanceTimersByTimeAsync(PROBE_MS);

    expect(signal.aborted).toBe(false);
  });
});
