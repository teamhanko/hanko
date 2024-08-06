import { Listener } from "../../../src/lib/events/Listener";
import {
  sessionCreatedType,
  sessionExpiredType,
  userDeletedType,
  userLoggedOutType,
} from "../../../src";

describe("Listener()", () => {
  let listener: Listener;
  let addEventListenerSpy: jest.SpyInstance;
  let mockThrottleFunc: jest.SpyInstance;
  let mockCallback: jest.Mock;

  beforeEach(() => {
    listener = new Listener();
    addEventListenerSpy = jest.spyOn(listener, "_addEventListener");
    mockThrottleFunc = jest.spyOn(listener, "_throttle");
    mockCallback = jest.fn();
    jest.useFakeTimers({ now: new Date() });
  });

  afterEach(() => {
    jest.restoreAllMocks();
    jest.clearAllTimers();
  });

  describe("onSessionCreated()", () => {
    it("should add an event listener for session created events", async () => {
      const mockDetail = {
        userID: "testUser",
        jwt: "testJWT",
        expirationSeconds: 7,
      };

      listener.onSessionCreated(mockCallback);

      expect(addEventListenerSpy).toHaveBeenCalledWith(
        sessionCreatedType,
        expect.any(Function),
        { once: false },
      );

      expect(mockThrottleFunc).toHaveBeenCalledWith(
        expect.any(Function),
        listener.throttleLimit,
        {
          leading: true,
          trailing: false,
        },
      );

      const mockEvent = new CustomEvent(sessionCreatedType, {
        detail: mockDetail,
      });

      // should throttle
      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);

      expect(mockCallback).toHaveBeenCalledWith(mockDetail);
      expect(mockCallback).toHaveBeenCalledTimes(1);
    });

    it("should only execute the callback once", async () => {
      const mockDetail = {
        userID: "testUser",
        jwt: "testJWT",
        expirationSeconds: 7,
      };

      const mockEvent = new CustomEvent(sessionCreatedType, {
        detail: mockDetail,
      });

      listener.onSessionCreated(mockCallback, true);

      expect(addEventListenerSpy).toHaveBeenCalledWith(
        sessionCreatedType,
        expect.any(Function),
        { once: true },
      );

      document.dispatchEvent(mockEvent);
      jest.advanceTimersByTime(1100); // skip throttle
      document.dispatchEvent(mockEvent);

      expect(mockCallback).toBeCalledTimes(1);
    });

    it("should clean up the event listener", async () => {
      const mockDetail = {
        userID: "testUser",
        jwt: "testJWT",
        expirationSeconds: 7,
      };

      const mockEvent = new CustomEvent(sessionCreatedType, {
        detail: mockDetail,
      });

      const cleanup = listener.onSessionCreated(mockCallback, true);

      cleanup();

      document.dispatchEvent(mockEvent);
      expect(mockCallback).toBeCalledTimes(0);
    });
  });

  describe("onSessionExpired()", () => {
    it("should add an event listener for session expired events", async () => {
      listener.onSessionExpired(mockCallback);

      expect(addEventListenerSpy).toHaveBeenCalledWith(
        sessionExpiredType,
        expect.any(Function),
        { once: false },
      );

      expect(mockThrottleFunc).toHaveBeenCalledWith(
        expect.any(Function),
        listener.throttleLimit,
        {
          leading: true,
          trailing: false,
        },
      );

      const mockEvent = new CustomEvent(sessionExpiredType, {});

      // should throttle
      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);

      expect(mockCallback).toHaveBeenCalledTimes(1);
    });

    it("should only execute the callback once", async () => {
      const mockEvent = new CustomEvent(sessionExpiredType, {});

      listener.onSessionExpired(mockCallback, true);

      expect(addEventListenerSpy).toHaveBeenCalledWith(
        sessionExpiredType,
        expect.any(Function),
        { once: true },
      );

      document.dispatchEvent(mockEvent);
      jest.advanceTimersByTime(1000); // skip throttle
      document.dispatchEvent(mockEvent);

      expect(mockCallback).toBeCalledTimes(1);
    });

    it("should clean up the event listener", async () => {
      const mockEvent = new CustomEvent(sessionExpiredType, {});

      const cleanup = listener.onSessionExpired(mockCallback, true);

      cleanup();

      document.dispatchEvent(mockEvent);
      expect(mockCallback).toBeCalledTimes(0);
    });
  });

  describe("onUserLogged()", () => {
    it("should add an event listener for user logged out events", async () => {
      listener.onUserLoggedOut(mockCallback);

      expect(addEventListenerSpy).toHaveBeenCalledWith(
        userLoggedOutType,
        expect.any(Function),
        { once: false },
      );

      expect(mockThrottleFunc).toBeCalledTimes(0);

      const mockEvent = new CustomEvent(userLoggedOutType, {});

      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);

      expect(mockCallback).toHaveBeenCalledTimes(3);
    });

    it("should only execute the callback once", async () => {
      const mockEvent = new CustomEvent(userLoggedOutType, {});

      listener.onUserLoggedOut(mockCallback, true);

      expect(addEventListenerSpy).toHaveBeenCalledWith(
        userLoggedOutType,
        expect.any(Function),
        { once: true },
      );

      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);

      expect(mockCallback).toBeCalledTimes(1);
    });

    it("should clean up the event listener", async () => {
      const mockEvent = new CustomEvent(userLoggedOutType, {});

      const cleanup = listener.onUserLoggedOut(mockCallback, true);

      cleanup();

      document.dispatchEvent(mockEvent);
      expect(mockCallback).toBeCalledTimes(0);
    });
  });

  describe("onUserDeleted()", () => {
    it("should add an event listener for user deleted events", async () => {
      listener.onUserDeleted(mockCallback);

      expect(addEventListenerSpy).toHaveBeenCalledWith(
        userDeletedType,
        expect.any(Function),
        { once: false },
      );

      expect(mockThrottleFunc).toBeCalledTimes(0);

      const mockEvent = new CustomEvent(userDeletedType, {});

      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);

      expect(mockCallback).toHaveBeenCalledTimes(3);
    });

    it("should only execute the callback once", async () => {
      const mockEvent = new CustomEvent(userDeletedType, {});

      listener.onUserDeleted(mockCallback, true);

      expect(addEventListenerSpy).toHaveBeenCalledWith(
        userDeletedType,
        expect.any(Function),
        { once: true },
      );

      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);

      expect(mockCallback).toBeCalledTimes(1);
    });

    it("should clean up the event listener", async () => {
      const mockEvent = new CustomEvent(userDeletedType, {});

      const cleanup = listener.onUserDeleted(mockCallback, true);

      cleanup();

      document.dispatchEvent(mockEvent);
      expect(mockCallback).toBeCalledTimes(0);
    });
  });
});
