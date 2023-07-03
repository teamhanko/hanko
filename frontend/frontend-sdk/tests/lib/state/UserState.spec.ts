import { decodedLSContent } from "../../setup";
import { UserState } from "../../../src/lib/state/users/UserState";

describe("userState.getUserState()", () => {
  it("should return the user state when local storage is initialized", async () => {
    const ls = decodedLSContent();
    const state = new (class extends UserState {})("hanko");
    const userID = Object.keys(decodedLSContent().users)[0];

    state.ls = decodedLSContent();
    expect(state.getUserState(userID)).toEqual(ls.users[userID]);
  });

  it("should return the user state when local storage is uninitialized", async () => {
    const ls = decodedLSContent();
    const state = new (class extends UserState {})("hanko");
    const userID = Object.keys(ls.users)[0];

    state.ls = {};
    expect(state.getUserState(userID)).toEqual({});
  });
});
