import { autoSteps } from "../../../src/lib/flow-api/auto-steps";
import * as Pkce from "../../../src/lib/Pkce";

jest.mock("../../../src/lib/Pkce", () => ({
  getStoredCodeVerifier: jest.fn(),
  clearStoredCodeVerifier: jest.fn(),
}));

// eslint-disable-next-line require-jsdoc
function createState(overrides: any = {}) {
  return {
    error: undefined,
    isCached: false,
    payload: { redirect_url: "https://example.com/redirect" },
    actions: {
      exchange_token: { run: jest.fn().mockResolvedValue({ name: "success" }) },
      back: {
        run: jest.fn().mockResolvedValue({
          name: "login_init",
          dispatchAfterStateChangeEvent: jest.fn(),
        }),
      },
    },
    saveToLocalStorage: jest.fn(),
    dispatchAfterStateChangeEvent: jest.fn(),
    ...overrides,
  };
}

describe("autoSteps.thirdparty", () => {
  const realLocation = window.location;

  beforeEach(() => {
    jest.clearAllMocks();

    // @ts-ignore
    delete window.location;
    // @ts-ignore
    window.location = {
      search: "",
      pathname: "/callback",
      assign: jest.fn(),
    };

    jest.spyOn(history, "replaceState").mockImplementation(() => {});
  });

  afterEach(() => {
    window.location = realLocation;
  });

  it("exchanges the token when hanko_token is present", async () => {
    window.location.search = "?hanko_token=abc123";
    (Pkce.getStoredCodeVerifier as jest.Mock).mockReturnValue("verifier-value");
    const state = createState();

    const result = await autoSteps.thirdparty(state as any);

    expect(state.actions.exchange_token.run).toHaveBeenCalledWith({
      token: "abc123",
      code_verifier: "verifier-value",
    });
    expect(Pkce.clearStoredCodeVerifier).toHaveBeenCalled();
    expect(history.replaceState).toHaveBeenCalledWith(null, null, "/callback");
    expect(result).toEqual({ name: "success" });
    expect(window.location.assign).not.toHaveBeenCalled();
  });

  it("falls back to undefined code_verifier when none is stored", async () => {
    window.location.search = "?hanko_token=abc123";
    (Pkce.getStoredCodeVerifier as jest.Mock).mockReturnValue(null);
    const state = createState();

    await autoSteps.thirdparty(state as any);

    expect(state.actions.exchange_token.run).toHaveBeenCalledWith({
      token: "abc123",
      code_verifier: undefined,
    });
  });

  it("maps an access_denied error query param to third_party_access_denied", async () => {
    window.location.search =
      "?error=access_denied&error_description=user%20cancelled";
    const state = createState();

    const result = await autoSteps.thirdparty(state as any);

    expect(state.actions.back.run).toHaveBeenCalledWith(null, {
      dispatchAfterStateChangeEvent: false,
    });
    expect(result.error).toEqual({
      code: "third_party_access_denied",
      message: "user cancelled",
    });
    expect(result.dispatchAfterStateChangeEvent).toHaveBeenCalled();
    expect(window.location.assign).not.toHaveBeenCalled();
  });

  it("maps any other error query param to technical_error", async () => {
    window.location.search = "?error=server_error&error_description=boom";
    const state = createState();

    const result = await autoSteps.thirdparty(state as any);

    expect(result.error).toEqual({
      code: "technical_error",
      message: "boom",
    });
  });

  it("propagates state.error back instead of redirecting with an undefined url", async () => {
    const state = createState({
      error: { code: "some_backend_error", message: "something went wrong" },
      payload: { redirect_url: undefined },
    });

    const result = await autoSteps.thirdparty(state as any);

    expect(state.actions.back.run).toHaveBeenCalledWith(null, {
      dispatchAfterStateChangeEvent: false,
    });
    expect(result.error).toEqual({
      code: "some_backend_error",
      message: "something went wrong",
    });
    expect(result.dispatchAfterStateChangeEvent).toHaveBeenCalled();
    expect(window.location.assign).not.toHaveBeenCalled();
  });

  it("redirects to the payload's redirect_url when there is no error and state is not cached", async () => {
    const state = createState({ isCached: false });

    const result = await autoSteps.thirdparty(state as any);

    expect(state.saveToLocalStorage).toHaveBeenCalled();
    expect(window.location.assign).toHaveBeenCalledWith(
      "https://example.com/redirect",
    );
    expect(result).toBe(state);
  });

  it("goes back instead of redirecting when the state is cached", async () => {
    const state = createState({ isCached: true });

    await autoSteps.thirdparty(state as any);

    expect(state.actions.back.run).toHaveBeenCalledWith();
    expect(state.saveToLocalStorage).not.toHaveBeenCalled();
    expect(window.location.assign).not.toHaveBeenCalled();
  });
});
