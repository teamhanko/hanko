import { Listener } from "../../../src/lib/events/Listener";
import {
  authFlowCompletedType,
  sessionCreatedType,
  sessionRemovedType,
  userDeletedType,
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
        { once: false }
      );

      expect(mockThrottleFunc).toHaveBeenCalledWith(
        expect.any(Function),
        listener.throttleLimit,
        {
          leading: true,
          trailing: false,
        }
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
        { once: true }
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

  describe("onSessionRemoved()", () => {
    it("should add an event listener for session removed events", async () => {
      listener.onSessionRemoved(mockCallback);

      expect(addEventListenerSpy).toHaveBeenCalledWith(
        sessionRemovedType,
        expect.any(Function),
        { once: false }
      );

      expect(mockThrottleFunc).toHaveBeenCalledWith(
        expect.any(Function),
        listener.throttleLimit,
        {
          leading: true,
          trailing: false,
        }
      );

      const mockEvent = new CustomEvent(sessionRemovedType, {});

      // should throttle
      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);

      expect(mockCallback).toHaveBeenCalledTimes(1);
    });

    it("should only execute the callback once", async () => {
      const mockEvent = new CustomEvent(sessionRemovedType, {});

      listener.onSessionRemoved(mockCallback, true);

      expect(addEventListenerSpy).toHaveBeenCalledWith(
        sessionRemovedType,
        expect.any(Function),
        { once: true }
      );

      document.dispatchEvent(mockEvent);
      jest.advanceTimersByTime(1000); // skip throttle
      document.dispatchEvent(mockEvent);

      expect(mockCallback).toBeCalledTimes(1);
    });

    it("should clean up the event listener", async () => {
      const mockEvent = new CustomEvent(sessionRemovedType, {});

      const cleanup = listener.onSessionRemoved(mockCallback, true);

      cleanup();

      document.dispatchEvent(mockEvent);
      expect(mockCallback).toBeCalledTimes(0);
    });
  });

  describe("onAuthFlowCompleted()", () => {
    it("should add an event listener for auth flow completed events", async () => {
      const mockDetail = {
        userID: "testUser",
      };

      listener.onAuthFlowCompleted(mockCallback);

      expect(addEventListenerSpy).toHaveBeenCalledWith(
        authFlowCompletedType,
        expect.any(Function),
        { once: false }
      );

      expect(mockThrottleFunc).toBeCalledTimes(0);

      const mockEvent = new CustomEvent(authFlowCompletedType, {
        detail: mockDetail,
      });

      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);

      expect(mockCallback).toHaveBeenCalledWith(mockDetail);
      expect(mockCallback).toHaveBeenCalledTimes(3);
    });

    it("should only execute the callback once", async () => {
      const mockDetail = {
        userID: "testUser",
      };

      const mockEvent = new CustomEvent(authFlowCompletedType, {
        detail: mockDetail,
      });

      listener.onAuthFlowCompleted(mockCallback, true);

      expect(addEventListenerSpy).toHaveBeenCalledWith(
        authFlowCompletedType,
        expect.any(Function),
        { once: true }
      );

      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);

      expect(mockCallback).toBeCalledTimes(1);
    });

    it("should clean up the event listener", async () => {
      const mockDetail = {
        userID: "testUser",
      };

      const mockEvent = new CustomEvent(authFlowCompletedType, {
        detail: mockDetail,
      });

      const cleanup = listener.onAuthFlowCompleted(mockCallback, true);

      cleanup();

      document.dispatchEvent(mockEvent);
      expect(mockCallback).toBeCalledTimes(0);
    });
  });

  describe("onUserDeleted()", () => {
    it("should add an event listener for auth flow completed events", async () => {
      listener.onUserDeleted(mockCallback);

      expect(addEventListenerSpy).toHaveBeenCalledWith(
        userDeletedType,
        expect.any(Function),
        { once: false }
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
        { once: true }
      );

      document.dispatchEvent(mockEvent);
      document.dispatchEvent(mockEvent);

      expect(mockCallback).toBeCalledTimes(1);
    });

    it("should clean up the event listener", async () => {
      const mockEvent = new CustomEvent(authFlowCompletedType, {});

      const cleanup = listener.onUserDeleted(mockCallback, true);

      cleanup();

      document.dispatchEvent(mockEvent);
      expect(mockCallback).toBeCalledTimes(0);
    });
  });
});
