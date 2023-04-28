import { Dispatcher } from "../../../src/lib/events/Dispatcher";
import {
  AuthFlowCompletedEventDetail,
  CustomEventWithDetail,
  SessionCreatedEventDetail,
} from "../../../src/lib/events/CustomEvents";

describe("Dispatcher", () => {
  let dispatcher: Dispatcher;

  beforeEach(() => {
    dispatcher = new Dispatcher();
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
        .calls[0][0] as CustomEventWithDetail<SessionCreatedEventDetail>;
      expect(event.type).toEqual("hanko-session-created");
      expect(event.detail).toBe(detail);
    });
  });

  describe("dispatchSessionRemovedEvent()", () => {
    it("dispatches a custom event with the 'hanko-session-removed' type and null detail", () => {
      const dispatchEventSpy = jest.spyOn(dispatcher, "_dispatchEvent");

      dispatcher.dispatchSessionRemovedEvent();

      expect(dispatchEventSpy).toHaveBeenCalledTimes(1);
      expect(dispatchEventSpy).toHaveBeenCalledWith(
        new CustomEventWithDetail("hanko-session-removed", null)
      );
      const event = dispatchEventSpy.mock
        .calls[0][0] as CustomEventWithDetail<null>;
      expect(event.type).toEqual("hanko-session-removed");
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
        .calls[0][0] as CustomEventWithDetail<AuthFlowCompletedEventDetail>;
      expect(event.type).toEqual("hanko-auth-flow-completed");
      expect(event.detail).toBe(detail);
    });
  });
});
