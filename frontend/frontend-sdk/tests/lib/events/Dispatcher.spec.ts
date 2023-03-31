import { Dispatcher } from "../../../src/lib/events/Dispatcher";
import { CustomEventWithDetail } from "../../../src/lib/events/CustomEvents";

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
    });
  });
});
