import { autoSteps } from "../../../src/lib/flow-api/auto-steps";

// Regression tests for the "thirdparty" auto-step. When the flow re-enters the
// thirdparty state without a redirect_url in its payload (e.g. after a failed
// token exchange, once hanko_token was already stripped from the URL), the SDK
// must not call window.location.assign(undefined), which navigates to
// "/undefined" (see issue #2805).
describe("autoSteps.thirdparty", () => {
  let assignMock: jest.Mock;
  const originalLocation = window.location;

  const buildState = (overrides: Record<string, unknown> = {}) => ({
    isCached: false,
    payload: {},
    saveToLocalStorage: jest.fn(),
    actions: {
      back: { run: jest.fn().mockResolvedValue({ name: "back_state" }) },
      exchange_token: { run: jest.fn() },
    },
    ...overrides,
  });

  beforeEach(() => {
    assignMock = jest.fn();
    Object.defineProperty(window, "location", {
      configurable: true,
      writable: true,
      value: { search: "", pathname: "/login", assign: assignMock },
    });
  });

  afterEach(() => {
    Object.defineProperty(window, "location", {
      configurable: true,
      writable: true,
      value: originalLocation,
    });
  });

  it("does not navigate when the payload has no redirect_url", async () => {
    const state = buildState({ isCached: false, payload: {} });

    await autoSteps.thirdparty(state as never);

    expect(assignMock).not.toHaveBeenCalled();
    expect(state.actions.back.run).toHaveBeenCalled();
  });

  it("navigates to the redirect_url when present and the state is not cached", async () => {
    const redirectUrl = "https://example.com/authorize";
    const state = buildState({
      isCached: false,
      payload: { redirect_url: redirectUrl },
    });

    await autoSteps.thirdparty(state as never);

    expect(state.saveToLocalStorage).toHaveBeenCalled();
    expect(assignMock).toHaveBeenCalledWith(redirectUrl);
    expect(state.actions.back.run).not.toHaveBeenCalled();
  });

  it("runs the back action when the state is cached", async () => {
    const state = buildState({
      isCached: true,
      payload: { redirect_url: "https://example.com/authorize" },
    });

    await autoSteps.thirdparty(state as never);

    expect(assignMock).not.toHaveBeenCalled();
    expect(state.actions.back.run).toHaveBeenCalled();
  });
});
