import { LocalStorage } from "../src/lib/state/State";

export const encodedLSContent = () =>
  "JTI1N0IlMjUyMnVzZXJzJTI1MjIlMjUzQSUyNTdCJTI1MjIxMzQ4ZDllNC1jNDE5LTQxNmEtYjMxZS1mMDUzOThhOGVhOWElMjUyMiUyNTNBJTI1N0IlMjUyMnBhc3Njb2RlJTI1MjIlMjUzQSUyNTdCJTI1MjJpZCUyNTIyJTI1M0ElMjUyMmIwNDVlZTZiLWE3OTAtNDcxZi1iMGViLTFkODYwZDJkYzYyNyUyNTIyJTI1MkMlMjUyMnR0bCUyNTIyJTI1M0ExNjY0MzgwMDAwJTI1MkMlMjUyMnJlc2VuZEFmdGVyJTI1MjIlMjUzQTE2NjQzODAwMDAlMjU3RCUyNTJDJTI1MjJwYXNzd29yZCUyNTIyJTI1M0ElMjU3QiUyNTIycmV0cnlBZnRlciUyNTIyJTI1M0ExNjY0MzgwMDAwJTI1N0QlMjUyQyUyNTIyd2ViYXV0aG4lMjUyMiUyNTNBJTI1N0IlMjUyMmNyZWRlbnRpYWxzJTI1MjIlMjUzQSUyNTVCJTI1MjI3bUZaM2VvSGNCcWpJaUFPRjMwc3VtVXNtcmhLUDhmV1dNdWxHcnhfdjkwZm5mQld2LTFIekFiaGVYWFg2MllwWmx1MG4zTnZoNGRqUlV5WFlvWEFmOXU0bWhaeGN2VFdxNkFzWHZKWEVRZVFEcmJHYVVUN29td0U4VktmRm5vJTI1MjIlMjU1RCUyNTdEJTI1N0QlMjU3RCUyNTJDJTI1MjJzZXNzaW9uJTI1MjIlMjUzQSUyNTdCJTI1MjJ1c2VySUQlMjUyMiUyNTNBJTI1MjJ0ZXN0LXVzZXIlMjUyMiUyNTJDJTI1MjJqd3QlMjUyMiUyNTNBJTI1MjJ0ZXN0LWp3dCUyNTIyJTI1MkMlMjUyMmV4cGlyeSUyNTIyJTI1M0ExNjY0MzgwMDAwJTI1N0QlMjU3RA==";

export const decodedLSContent = (): LocalStorage => ({
  users: {
    "1348d9e4-c419-416a-b31e-f05398a8ea9a": {
      passcode: {
        id: "b045ee6b-a790-471f-b0eb-1d860d2dc627",
        ttl: 1664380000,
        resendAfter: 1664380000,
      },
      password: {
        retryAfter: 1664380000,
      },
      webauthn: {
        credentials: [
          "7mFZ3eoHcBqjIiAOF30sumUsmrhKP8fWWMulGrx_v90fnfBWv-1HzAbheXXX62YpZlu0n3Nvh4djRUyXYoXAf9u4mhZxcvTWq6AsXvJXEQeQDrbGaUT7omwE8VKfFno",
        ],
      },
    },
  },
  session: {
    userID: "test-user",
    jwt: "test-jwt",
    expiry: 1664380000,
  },
});

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

jest.useFakeTimers({
  now: 1664379699000,
});
