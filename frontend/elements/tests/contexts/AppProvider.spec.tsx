import { h, render } from "preact";
import { act } from "preact/test-utils";
import type { Hanko, State } from "@teamhanko/hanko-frontend-sdk";

import AppProvider from "../../src/contexts/AppProvider";

jest.mock("@denysvuika/preact-translate", () => ({
  TranslateProvider: ({ children }: { children: unknown }) => children,
}));

jest.mock("@teamhanko/hanko-frontend-sdk", () => ({
  HankoError: class extends Error {},
  TechnicalError: class extends Error {},
}));

jest.mock(
  "../../src/components/wrapper/Container",
  (): { __esModule: true; default: unknown } => {
    const { h: createElement } = jest.requireActual("preact");
    const { forwardRef } = jest.requireActual("preact/compat");

    return {
      __esModule: true,
      default: forwardRef(
        ({ children }: { children: unknown }, ref: { current: HTMLElement }) =>
          createElement("section", { ref }, children),
      ),
    };
  },
);

jest.mock("../../src/pages/InitPage", () => ({
  __esModule: true,
  default: (): null => null,
}));

jest.mock("../../src/pages/PasscodePage", () => ({
  __esModule: true,
  default: (): null => null,
}));

type EventCallback = () => void;
type RestartEvent = "onSessionExpired" | "onUserDeleted" | "onUserLoggedOut";
type StateChangeCallback = (detail: { state: State }) => void;

const AUTH_COMPONENT = "auth";
const LOGIN_FLOW = "login";
const ACTIVE_FLOW_STATE = "passcode_confirmation";
const COMPLETED_FLOW_STATE = "success";
const INITIALIZATION_DELAY_MS = 50;
const INITIAL_CREATE_STATE_CALLS = 1;
const REINITIALIZED_CREATE_STATE_CALLS = 2;

describe("AppProvider session events", () => {
  let container: HTMLDivElement;
  let eventCallbacks: Record<RestartEvent, EventCallback[]>;
  let stateChangeCallbacks: StateChangeCallback[];
  let createState: jest.Mock;
  let hanko: Hanko;

  const renderProvider = () => {
    act(() => {
      render(
        <AppProvider
          componentName={AUTH_COMPONENT}
          globalOptions={{
            hanko,
            injectStyles: false,
            fallbackLanguage: "en",
            storageKey: "hanko",
          }}
          createWebauthnAbortSignal={() => new AbortController().signal}
        />,
        container,
      );
    });
  };

  const waitForInitialization = async () => {
    await act(async () => {
      await new Promise((resolve) =>
        setTimeout(resolve, INITIALIZATION_DELAY_MS),
      );
    });
  };

  const dispatchStateChange = (
    name: typeof ACTIVE_FLOW_STATE | typeof COMPLETED_FLOW_STATE,
  ) => {
    const state = {
      name,
      flowName: LOGIN_FLOW,
      payload: null,
      autoStep: jest.fn().mockResolvedValue(undefined),
    } as unknown as State;

    act(() => {
      stateChangeCallbacks.forEach((callback) => callback({ state }));
    });
  };

  beforeEach(() => {
    eventCallbacks = {
      onSessionExpired: [],
      onUserDeleted: [],
      onUserLoggedOut: [],
    };
    stateChangeCallbacks = [];
    container = document.createElement("div");
    document.body.appendChild(container);

    const subscribe = () => jest.fn();
    const subscribeToRestartEvent = (eventName: RestartEvent) =>
      jest.fn((callback: EventCallback) => {
        eventCallbacks[eventName].push(callback);
        return jest.fn();
      });
    createState = jest.fn().mockResolvedValue({
      flowName: LOGIN_FLOW,
      dispatchAfterStateChangeEvent: jest.fn(),
    });
    hanko = {
      setLang: jest.fn(),
      createState,
      onBeforeStateChange: jest.fn(subscribe),
      onAfterStateChange: jest.fn((callback: StateChangeCallback) => {
        stateChangeCallbacks.push(callback);
        return jest.fn();
      }),
      onSessionCreated: jest.fn(subscribe),
      onUserLoggedOut: subscribeToRestartEvent("onUserLoggedOut"),
      onUserDeleted: subscribeToRestartEvent("onUserDeleted"),
      onSessionExpired: subscribeToRestartEvent("onSessionExpired"),
    } as unknown as Hanko;
  });

  afterEach(() => {
    render(null, container);
    container.remove();
  });

  it("keeps an active authentication flow when a stale session expires", async () => {
    renderProvider();
    await waitForInitialization();
    dispatchStateChange(ACTIVE_FLOW_STATE);

    expect(createState).toHaveBeenCalledTimes(INITIAL_CREATE_STATE_CALLS);

    await act(async () => {
      eventCallbacks.onSessionExpired.forEach((callback) => callback());
      await Promise.resolve();
    });

    expect(createState).toHaveBeenCalledTimes(INITIAL_CREATE_STATE_CALLS);
  });

  it("restarts an inactive authentication flow after session expiry", async () => {
    renderProvider();
    await waitForInitialization();
    dispatchStateChange(COMPLETED_FLOW_STATE);

    eventCallbacks.onSessionExpired.forEach((callback) => callback());

    expect(createState).toHaveBeenCalledTimes(REINITIALIZED_CREATE_STATE_CALLS);
  });

  it.each([
    ["user logout", "onUserLoggedOut"],
    ["user deletion", "onUserDeleted"],
  ] as const)(
    "restarts the authentication flow after %s",
    async (_, eventName) => {
      renderProvider();
      await waitForInitialization();

      eventCallbacks[eventName].forEach((callback) => callback());

      expect(createState).toHaveBeenCalledTimes(
        REINITIALIZED_CREATE_STATE_CALLS,
      );
    },
  );
});
