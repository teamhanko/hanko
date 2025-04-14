import {
  State,
  Action,
  Hanko,
  FlowResponse,
  AnyState,
  FlowName,
} from "../../../src";
import { autoSteps } from "../../../src/lib/flow-api/auto-steps";

jest.mock("../../../src", () => {
  const actual = jest.requireActual("../../../src");
  return {
    ...actual,
    Hanko: jest.fn(() => ({
      client: {
        post: jest.fn<() => Promise<Response>, [string, any?]>(),
      },
      relay: {
        dispatchAfterStateChangeEvent: jest.fn(),
        dispatchBeforeStateChangeEvent: jest.fn(),
      },
    })),
  };
});

jest.mock("../../../src/lib/flow-api/auto-steps", () => ({
  autoSteps: {
    preflight: jest.fn(),
  },
}));

jest.mock("../../../src/lib/flow-api/passkey-autofill-activation", () => ({
  passkeyAutofillActivationHandlers: {
    somePasskeyState: jest.fn(),
  },
}));

describe("State", () => {
  let hankoMock: jest.Mocked<Hanko>;
  const flowName: FlowName = "login";
  const defaultCacheKey = "hanko-flow-state"; // Updated to match code's default
  const mockLoginInitResponse: FlowResponse<"login_init"> = {
    name: "login_init",
    csrf_token: "csrf123",
    status: 200,
    actions: {
      continue_with_login_identifier: {
        action: "continue_with_login_identifier",
        href: "/continue_with_login_identifier",
        inputs: {},
        description: null,
      },
    },
    payload: null,
  };
  const mockPreflightResponse: FlowResponse<"preflight"> = {
    ...mockLoginInitResponse,
    name: "preflight",
    payload: null,
    actions: {
      register_client_capabilities: {
        action: "register_client_capabilities",
        href: "/register_client_capabilities",
        inputs: {
          webauthn_available: {
            name: "webauthn_available",
            type: "boolean",
          },
          webauthn_conditional_mediation_available: {
            name: "webauthn_conditional_mediation_available",
            type: "boolean",
          },
          webauthn_platform_authenticator_available: {
            name: "webauthn_platform_authenticator_available",
            type: "boolean",
          },
        },
        description: "",
      },
    },
  };

  beforeEach(() => {
    hankoMock = new (jest.requireMock(
      "../../../src",
    ).Hanko)() as jest.Mocked<Hanko>;
    localStorage.clear();
    jest.clearAllMocks();
  });

  describe("constructor", () => {
    it("initializes state properties correctly", () => {
      const state = new State(hankoMock, flowName, mockLoginInitResponse);
      expect(state.name).toBe("login_init");
      expect(state.flowName).toBe(flowName);
      expect(state.csrfToken).toBe("csrf123");
      expect(state.status).toBe(200);
      expect(state.payload).toBeNull();
      expect(state.hanko).toBe(hankoMock);
      expect(state.actions.continue_with_login_identifier).toBeInstanceOf(
        Action,
      );
      expect(state.isCached).toBe(false);
      expect(state.cacheKey).toBe("hanko-flow-state");
    });

    it("sets up autoStep if state is in autoSteps", () => {
      const state = new State(hankoMock, flowName, mockPreflightResponse);
      expect(state.autoStep).toBeDefined();
      expect(typeof state.autoStep).toBe("function");
    });

    it("dispatches state change event by default", () => {
      new State(hankoMock, flowName, mockLoginInitResponse);
      expect(
        hankoMock.relay.dispatchAfterStateChangeEvent,
      ).toHaveBeenCalledWith({
        state: expect.any(State),
      });
    });
  });

  describe("buildActions", () => {
    it("creates actions map with proxy for undefined actions", () => {
      const state = new State(hankoMock, flowName, mockLoginInitResponse);
      expect(state.actions.continue_with_login_identifier).toBeInstanceOf(
        Action,
      );
      expect(state.actions.continue_with_login_identifier.enabled).toBe(true);
      expect(state.actions.thirdparty_oauth).toBeInstanceOf(Action);
      expect(state.actions.thirdparty_oauth.enabled).toBe(false);
    });
  });

  describe("saveToLocalStorage", () => {
    it("saves serialized state to localStorage", () => {
      const state = new State(hankoMock, flowName, mockLoginInitResponse);
      state.saveToLocalStorage();
      expect(localStorage.setItem).toHaveBeenCalled();
      const setItemCall = (localStorage.setItem as jest.Mock).mock.calls[0];
      const [key, value] = setItemCall;
      expect(key).toBe(defaultCacheKey);
      const parsedValue = JSON.parse(value);
      expect(parsedValue).toEqual({
        ...mockLoginInitResponse,
        flow_name: flowName,
        is_cached: true,
        previous_action: null,
      });
    });

    it("uses custom cacheKey for cached data", async () => {
      const customCacheKey = "custom-hanko-state";
      (localStorage.getItem as jest.Mock).mockReturnValue(
        JSON.stringify({ ...mockLoginInitResponse, is_cached: true }),
      );
      const state = await State.create(hankoMock, flowName, {
        cacheKey: customCacheKey,
      });
      expect(localStorage.getItem).toHaveBeenCalledWith(customCacheKey);
      expect(state.cacheKey).toBe(customCacheKey);
      expect(state.name).toBe("login_init");
    });
  });

  describe("static deserialize", () => {
    it("creates state from serialized data", async () => {
      const serializedState = {
        ...mockLoginInitResponse,
        flow_name: flowName,
        is_cached: true,
      };
      const state = await State.deserialize(hankoMock, serializedState);
      expect(state.name).toBe("login_init");
      expect(state.isCached).toBe(true);
      expect(state.flowName).toBe(flowName);
    });
  });

  describe("readFromLocalStorage", () => {
    it("returns undefined for invalid JSON", () => {
      (localStorage.getItem as jest.Mock).mockReturnValue("invalid-json");
      const result = State.readFromLocalStorage(defaultCacheKey);
      expect(result).toBeUndefined();
    });
  });

  describe("static create", () => {
    it("creates state from cached data if available", async () => {
      (localStorage.getItem as jest.Mock).mockReturnValue(
        JSON.stringify({ ...mockLoginInitResponse, is_cached: true }),
      );
      const state = await State.create(hankoMock, flowName);
      expect(hankoMock.client.post).not.toHaveBeenCalled();
      expect(state.name).toBe("login_init");
      expect(state.isCached).toBe(true);
    });

    it("fetches state if no cached data", async () => {
      (localStorage.getItem as jest.Mock).mockReturnValue(null);
      (hankoMock.client.post as jest.Mock).mockResolvedValue({
        json: () => Promise.resolve(mockLoginInitResponse),
      });
      const state = await State.create(hankoMock, flowName);
      expect(hankoMock.client.post).toHaveBeenCalled();
      expect(state.name).toBe("login_init");
      expect(state.isCached).toBe(false);
    });

    it("respects loadFromCache: false", async () => {
      (localStorage.getItem as jest.Mock).mockReturnValue(
        JSON.stringify({ ...mockLoginInitResponse, is_cached: true }),
      );
      (hankoMock.client.post as jest.Mock).mockResolvedValue({
        json: () => Promise.resolve(mockLoginInitResponse),
      });
      const state = await State.create(hankoMock, flowName, {
        loadFromCache: false,
      });
      expect(hankoMock.client.post).toHaveBeenCalled();
      expect(state.name).toBe("login_init");
      expect(state.isCached).toBe(false);
    });
  });

  describe("static fetchState", () => {
    it("returns error response on fetch failure", async () => {
      (hankoMock.client.post as jest.Mock).mockRejectedValue(
        new Error("Network error"),
      );
      const response = await State.fetchState(
        hankoMock,
        "/continue_with_login_identifier",
      );
      expect(response.name).toBe("error");
      expect(response.error).toBeDefined();
    });
  });

  describe("initializeFlowState", () => {
    it("processes auto-steps if not excluded", async () => {
      const nextState = { name: "preflight" } as AnyState;
      (autoSteps.preflight as jest.Mock).mockResolvedValue(nextState);
      const state = await State.initializeFlowState(
        hankoMock,
        flowName,
        mockPreflightResponse,
      );
      expect(autoSteps.preflight).toHaveBeenCalled();
      expect(state.name).toBe("preflight");
    });
  });
});

describe("Action", () => {
  let hankoMock: jest.Mocked<Hanko>;
  let state: State<"login_init">;
  const flowName: FlowName = "login";
  const mockResponse: FlowResponse<"login_init"> = {
    name: "login_init",
    csrf_token: "csrf123",
    status: 200,
    actions: {
      continue_with_login_identifier: {
        action: "continue_with_login_identifier",
        href: "/continue_with_login_identifier",
        inputs: {},
        description: "",
      },
    },
    payload: null,
  };
  const actionDef = {
    action: "continue_with_login_identifier",
    href: "/continue_with_login_identifier",
    inputs: { username: { value: "default" } },
    description: "",
  };

  beforeEach(() => {
    hankoMock = new (jest.requireMock(
      "../../../src",
    ).Hanko)() as jest.Mocked<Hanko>;
    state = new State(hankoMock, flowName, mockResponse);
    localStorage.clear();
    jest.clearAllMocks();
  });

  describe("constructor", () => {
    it("initializes action properties", () => {
      const action = new Action(actionDef, state);
      expect(action.name).toBe("continue_with_login_identifier");
      expect(action.href).toBe("/continue_with_login_identifier");
      expect(action.enabled).toBe(true);
      expect(action.inputs).toEqual({ username: { value: "default" } });
    });
  });

  describe("createDisabled", () => {
    it("creates a disabled action", () => {
      const action = Action.createDisabled("disabledAction", state);
      expect(action.enabled).toBe(false);
      expect(action.name).toBe("disabledAction");
      expect(action.href).toBe("");
    });
  });

  describe("run", () => {
    it("executes action and returns new state", async () => {
      const action = new Action(actionDef, state);
      const nextResponse: FlowResponse<any> = {
        ...mockResponse,
        name: "nextState",
      };
      (hankoMock.client.post as jest.Mock).mockResolvedValue({
        json: () => Promise.resolve(nextResponse),
      });
      // @ts-ignore
      const newState = await action.run({ username: "custom" });
      expect(hankoMock.client.post).toHaveBeenCalledWith(
        "/continue_with_login_identifier",
        {
          input_data: { username: "custom" },
          csrf_token: "csrf123",
        },
      );
      expect(newState.name).toBe("nextState");
    });

    it("throws if action is disabled", async () => {
      const action = Action.createDisabled("disabledAction", state);
      await expect(action.run()).rejects.toThrow(
        "Action 'disabledAction' is not enabled",
      );
    });

    it("throws if action already invoked", async () => {
      const action = new Action(actionDef, state);
      state.invokedAction = {
        name: "previousAction",
        relatedStateName: "login_init",
      };
      await expect(action.run()).rejects.toThrow(
        "An action 'previousAction' has already been invoked on state 'login_init'. No further actions can be run.",
      );
    });
  });
});
