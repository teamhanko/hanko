import { decodedLSContent } from "../../setup";
import { SessionState } from "../../../src/lib/state/session/SessionState";

describe("sessionState.read()", () => {
  it("should read the session state", async () => {
    const state = new SessionState({ localStorageKey: "hanko" });

    expect(state.read()).toEqual(state);
  });
});

describe("sessionState.reset()", () => {
  it("should reset information about the current session", async () => {
    const ls = decodedLSContent();
    const state = new SessionState({ localStorageKey: "hanko" });

    state.ls = ls;

    expect(state.reset()).toEqual(state);
    expect(state.ls.session.expiry).toBeUndefined();
  });
});

describe("sessionState.getExpirationSeconds()", () => {
  it("should return seconds until the session is active", async () => {
    const ls = decodedLSContent();
    const state = new SessionState({ localStorageKey: "hanko" });

    state.ls = ls;

    expect(state.getExpirationSeconds()).toEqual(301);
  });
});

describe("sessionState.setExpirationSeconds()", () => {
  it("should set a timestamp until the session is active", async () => {
    const state = new SessionState({ localStorageKey: "hanko" });
    const seconds = 42;

    expect(state.setExpirationSeconds(seconds)).toEqual(state);
    expect(state.ls.session.expiry).toEqual(
      Math.floor(Date.now() / 1000) + seconds,
    );
  });
});
