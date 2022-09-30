import { decodedLSContent } from "../../setup";
import { PasscodeState } from "../../../src/lib/state/PasscodeState";

describe("passcodeState.read()", () => {
  it("should read the password state", async () => {
    const state = new PasscodeState();

    expect(state.read()).toEqual(state);
  });
});

describe("passcodeState.getActiveID()", () => {
  it("should return the id of the currently active passcode", async () => {
    const ls = decodedLSContent();
    const state = new PasscodeState();
    const userID = Object.keys(ls.users)[0];

    state.ls = ls;

    expect(state.getActiveID(userID)).toEqual(ls.users[userID].passcode.id);
  });
});

describe("passcodeState.setActiveID()", () => {
  it("should return the id of the currently active passcode", async () => {
    const ls = decodedLSContent();
    const state = new PasscodeState();
    const userID = Object.keys(ls.users)[0];
    const passcodeID = "test_id_1";

    state.ls = ls;

    expect(state.setActiveID(userID, passcodeID)).toEqual(state);
    expect(state.ls.users[userID].passcode.id).toEqual(passcodeID);
  });
});

describe("passcodeState.reset()", () => {
  it("should return the id of the currently active passcode", async () => {
    const ls = decodedLSContent();
    const state = new PasscodeState();
    const userID = Object.keys(ls.users)[0];

    state.ls = ls;

    expect(state.reset(userID)).toEqual(state);
    expect(state.ls.users[userID].passcode.id).toBeUndefined();
    expect(state.ls.users[userID].passcode.ttl).toBeUndefined();
    expect(state.ls.users[userID].passcode.resendAfter).toBeUndefined();
  });
});

describe("passcodeState.getResendAfter()", () => {
  it("should return seconds until a new passcode can be send", async () => {
    const ls = decodedLSContent();
    const state = new PasscodeState();
    const userID = Object.keys(decodedLSContent().users)[0];

    state.ls = ls;

    expect(state.getResendAfter(userID)).toEqual(301);
  });
});

describe("passcodeState.setResendAfter()", () => {
  it("should set a timestamp until a new passcode can be send", async () => {
    const ls = decodedLSContent();
    const state = new PasscodeState();
    const userID = Object.keys(ls.users)[0];
    const seconds = 42;

    expect(state.setResendAfter(userID, seconds)).toEqual(state);
    expect(state.ls.users[userID].passcode.resendAfter).toEqual(
      Math.floor(Date.now() / 1000) + seconds
    );
  });
});

describe("passcodeState.getTTL()", () => {
  it("should return seconds until the active passcode lives", async () => {
    const ls = decodedLSContent();
    const state = new PasscodeState();
    const userID = Object.keys(ls.users)[0];

    state.ls = ls;

    expect(state.getTTL(userID)).toEqual(301);
  });
});

describe("passcodeState.setTTL()", () => {
  it("should set a timestamp until the active passcode lives", async () => {
    const ls = decodedLSContent();
    const state = new PasscodeState();
    const userID = Object.keys(ls.users)[0];
    const seconds = 42;

    expect(state.setTTL(userID, seconds)).toEqual(state);
    expect(state.ls.users[userID].passcode.ttl).toEqual(
      Math.floor(Date.now() / 1000) + seconds
    );
  });
});
