import { decodedLSContent } from "../../setup";
import { PasswordState } from "../../../src/lib/state/users/PasswordState";

describe("passwordState.read()", () => {
  it("should read the password state", async () => {
    const state = new PasswordState("hanko");

    expect(state.read()).toEqual(state);
  });
});

describe("passwordState.getRetryAfter()", () => {
  it("should return seconds until a new login can be attempted", async () => {
    const ls = decodedLSContent();
    const state = new PasswordState("hanko");
    const userID = Object.keys(ls.users)[0];

    state.ls = ls;

    expect(state.getRetryAfter(userID)).toEqual(301);
  });
});

describe("passwordState.setRetryAfter()", () => {
  it("should set a timestamp until a new login can be attempted", async () => {
    const ls = decodedLSContent();
    const state = new PasswordState("hanko");
    const userID = Object.keys(ls.users)[0];
    const seconds = 42;

    expect(state.setRetryAfter(userID, seconds)).toEqual(state);
    expect(state.ls.users[userID].password.retryAfter).toEqual(
      Math.floor(Date.now() / 1000) + seconds
    );
  });
});
