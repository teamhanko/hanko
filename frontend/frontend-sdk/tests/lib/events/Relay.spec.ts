import { Relay } from "../../../src/lib/events/Relay";
import {
  CustomEventWithDetail,
  sessionCreatedType,
  sessionExpiredType,
  userDeletedType,
} from "../../../src";

describe("Relay", () => {
  let relay: Relay;
  let dispatcherSpy: jest.SpyInstance;

  beforeEach(() => {
    relay = new Relay();
    dispatcherSpy = jest.spyOn(relay, "_dispatchEvent");
  });

  afterEach(() => {
    jest.clearAllMocks();
    jest.clearAllTimers();
  });

  it("should listen to 'hanko-session-created' events and remove scheduled events", () => {
    const mockSessionCreatedDetail = {
      userID: "test-user",
      jwt: "test-token",
      expirationSeconds: 7,
    };
    const sessionCreatedEventMock = new CustomEventWithDetail(
      sessionCreatedType,
      mockSessionCreatedDetail
    );
    const sessionExpiredEventMock = new CustomEventWithDetail(
      sessionExpiredType,
      null
    );

    document.dispatchEvent(sessionCreatedEventMock);

    expect(relay._scheduler._tasks).toStrictEqual([
      {
        func: expect.any(Function),
        timeoutID: expect.any(Number),
        type: sessionExpiredType,
      },
    ]);

    document.dispatchEvent(sessionExpiredEventMock);

    expect(relay._scheduler._tasks).toStrictEqual([]);

    jest.advanceTimersByTime(7000);
    expect(dispatcherSpy).toBeCalledTimes(0);
  });

  it("should listen to 'hanko-user-deleted' and remove scheduled events", () => {
    const mockSessionCreatedDetail = {
      userID: "test-user",
      jwt: "test-token",
      expirationSeconds: 7,
    };
    const sessionCreatedEventMock = new CustomEventWithDetail(
      sessionCreatedType,
      mockSessionCreatedDetail
    );
    const userDeletedEventMock = new CustomEventWithDetail(
      userDeletedType,
      null
    );

    document.dispatchEvent(sessionCreatedEventMock);

    expect(relay._scheduler._tasks).toStrictEqual([
      {
        func: expect.any(Function),
        timeoutID: expect.any(Number),
        type: sessionExpiredType,
      },
    ]);

    document.dispatchEvent(userDeletedEventMock);

    expect(relay._scheduler._tasks).toStrictEqual([]);

    jest.advanceTimersByTime(7000);

    expect(dispatcherSpy).toBeCalledTimes(0);
  });

  it("should listen to 'storage' events and dispatch 'hanko-session-expired' if the session is expired", () => {
    jest.spyOn(relay._session._sessionState, "getUserID").mockReturnValue("");
    jest.spyOn(relay._session._cookie, "getAuthCookie").mockReturnValue("");
    jest
      .spyOn(relay._session._sessionState, "getExpirationSeconds")
      .mockReturnValue(0);

    window.dispatchEvent(
      new StorageEvent("storage", {
        key: "hanko_session",
      })
    );

    expect(dispatcherSpy).toHaveBeenCalled();
    expect(dispatcherSpy.mock.calls[0][0].type).toEqual(sessionExpiredType);
  });

  it("should listen to 'storage' events and dispatch 'hanko-session-created' if session is active", () => {
    jest
      .spyOn(relay._session._sessionState, "getUserID")
      .mockReturnValue("test-user");
    jest
      .spyOn(relay._session._cookie, "getAuthCookie")
      .mockReturnValue("test-jwt");
    jest
      .spyOn(relay._session._sessionState, "getExpirationSeconds")
      .mockReturnValue(10);

    window.dispatchEvent(
      new StorageEvent("storage", {
        key: "hanko_session",
      })
    );

    expect(dispatcherSpy).toHaveBeenCalled();
    expect(dispatcherSpy.mock.calls[0][0].type).toEqual(sessionCreatedType);
  });
});
