const fakeLocalStorage = (function () {
  return {
    getItem: jest.fn(),
    setItem: jest.fn(),
    removeItem: jest.fn(),
    clear: jest.fn(),
  };
})();

Object.defineProperty(global, "localStorage", {
  value: fakeLocalStorage,
});

const fakeCredentials = (function () {
  return {
    create: jest.fn(),
    get: jest.fn(),
  };
})();

const fakeNavigator = (function () {
  return {
    credentials: fakeCredentials,
  };
})();

Object.defineProperty(global, "navigator", {
  value: fakeNavigator,
});

export const fakePublicKeyCredential = (function () {
  return {
    isUserVerifyingPlatformAuthenticatorAvailable: jest.fn(),
    isExternalCTAP2SecurityKeySupported: jest.fn(),
    isConditionalMediationAvailable: jest.fn(),
  };
})();

Object.defineProperty(window, "PublicKeyCredential", {
  value: fakePublicKeyCredential,
  configurable: true,
  writable: true,
});

export const fakeXMLHttpRequest = (function () {
  return jest.fn().mockImplementation(() => ({
    response: "{}",
    open: jest.fn(),
    setRequestHeader: jest.fn(),
    getResponseHeader: jest.fn(),
    send: jest.fn(),
  }));
})();

Object.defineProperty(global, "XMLHttpRequest", {
  value: fakeXMLHttpRequest,
  configurable: true,
  writable: true,
});

export const fakeTimerNow = 1664379699000;

jest.useFakeTimers({
  now: fakeTimerNow,
});

// eslint-disable-next-line require-jsdoc
class MockBroadcastChannel {
  name: string;
  message: any;

  constructor(name: string) {
    this.name = name;
  }

  postMessage(message: any) {
    this.message = message;
  }
}

Object.defineProperty(global, "BroadcastChannel", {
  value: MockBroadcastChannel,
  configurable: true,
  writable: true,
});
