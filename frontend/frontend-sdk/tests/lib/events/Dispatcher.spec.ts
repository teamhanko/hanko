import { Dispatcher } from "../../../src/lib/events/Dispatcher";
import {
  AuthFlowCompletedDetail,
  CustomEventWithDetail,
  SessionDetail,
} from "../../../src";

describe("Dispatcher", () => {
  let dispatcher: Dispatcher;

  beforeEach(() => {
    dispatcher = new Dispatcher("hanko");
  });

  describe("dispatchSessionCreatedEvent()", () => {
    it("dispatches a custom event with the 'hanko-session-created' type and the provided detail", () => {
      const detail = {
        userID: "test-user",
        jwt: "test-token",
        expirationSeconds: 7,
      };
      const dispatchEventSpy = jest.spyOn(dispatcher, "_dispatchEvent");

      dispatcher.dispatchSessionCreatedEvent(detail);

      expect(dispatchEventSpy).toHaveBeenCalledTimes(1);
      expect(dispatchEventSpy).toHaveBeenCalledWith(
        new CustomEventWithDetail("hanko-session-created", detail)
      );
      const event = dispatchEventSpy.mock
        .calls[0][0] as CustomEventWithDetail<SessionDetail>;
      expect(event.type).toEqual("hanko-session-created");
      expect(event.detail).toBe(detail);
    });
  });

  describe("dispatchSessionExpiredEvent()", () => {
    it("dispatches a custom event with the 'hanko-session-expired' type and null detail", () => {
      const dispatchEventSpy = jest.spyOn(dispatcher, "_dispatchEvent");

      dispatcher.dispatchSessionExpiredEvent();

      expect(dispatchEventSpy).toHaveBeenCalledTimes(1);
      expect(dispatchEventSpy).toHaveBeenCalledWith(
        new CustomEventWithDetail("hanko-session-expired", null)
      );
      const event = dispatchEventSpy.mock
        .calls[0][0] as CustomEventWithDetail<null>;
      expect(event.type).toEqual("hanko-session-expired");
    });
  });

  describe("dispatchUserDeletedEvent()", () => {
    it("dispatches a custom event with the 'hanko-user-deleted' type and null detail", () => {
      const dispatchEventSpy = jest.spyOn(dispatcher, "_dispatchEvent");

      dispatcher.dispatchUserDeletedEvent();

      expect(dispatchEventSpy).toHaveBeenCalledTimes(1);
      expect(dispatchEventSpy).toHaveBeenCalledWith(
        new CustomEventWithDetail("hanko-user-deleted", null)
      );
      const event = dispatchEventSpy.mock
        .calls[0][0] as CustomEventWithDetail<null>;
      expect(event.type).toEqual("hanko-user-deleted");
    });
  });

  describe("dispatchUserLoggedOutEvent()", () => {
    it("dispatches a custom event with the 'hanko-user-logged-out' type and null detail", () => {
      const dispatchEventSpy = jest.spyOn(dispatcher, "_dispatchEvent");

      dispatcher.dispatchUserLoggedOutEvent();

      expect(dispatchEventSpy).toHaveBeenCalledTimes(1);
      expect(dispatchEventSpy).toHaveBeenCalledWith(
        new CustomEventWithDetail("hanko-user-logged-out", null)
      );
      const event = dispatchEventSpy.mock
        .calls[0][0] as CustomEventWithDetail<null>;
      expect(event.type).toEqual("hanko-user-logged-out");
    });
  });

  describe("dispatchAuthFlowCompletedEvent()", () => {
    it("dispatches a custom event with the 'hanko-auth-flow-completed' type and the provided detail", () => {
      const detail = { userID: "test-user" };
      const dispatchEventSpy = jest.spyOn(dispatcher, "_dispatchEvent");

      dispatcher.dispatchAuthFlowCompletedEvent(detail);

      expect(dispatchEventSpy).toHaveBeenCalledTimes(1);
      expect(dispatchEventSpy).toHaveBeenCalledWith(
        new CustomEventWithDetail("hanko-auth-flow-completed", detail)
      );
      const event = dispatchEventSpy.mock
        .calls[0][0] as CustomEventWithDetail<AuthFlowCompletedDetail>;
      expect(event.type).toEqual("hanko-auth-flow-completed");
      expect(event.detail).toBe(detail);
    });
  });
});
