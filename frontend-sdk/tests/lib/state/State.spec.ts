import { State } from "../../../src/lib/state/State";
import { decodedLSContent, encodedLSContent } from "../../setup";

describe("state.read()", () => {
  it("should read when local storage contents are initialized", async () => {
    const ls = decodedLSContent();
    const state = new (class extends State {})();

    jest.spyOn(localStorage, "getItem").mockReturnValueOnce(encodedLSContent());

    expect(state.read()).toEqual(state);
    expect(state.ls).toEqual(ls);
    expect(localStorage.getItem).toHaveBeenCalledTimes(1);
    expect(localStorage.getItem).toHaveBeenCalledWith("hanko");
  });

  it("should read when local storage contents are corrupted", async () => {
    const state = new (class extends State {})();

    jest.spyOn(localStorage, "getItem").mockReturnValueOnce("junk");

    expect(state.read()).toEqual(state);
    expect(state.ls).toEqual({});
  });
});

describe("state.write()", () => {
  it("should write local storage contents", async () => {
    const ls = decodedLSContent();
    const state = new (class extends State {})();

    state.ls = ls;
    jest.spyOn(localStorage, "setItem");

    expect(state.write()).toEqual(state);
    expect(localStorage.setItem).toHaveBeenCalledWith(
      "hanko",
      encodedLSContent()
    );
    expect(localStorage.setItem).toHaveBeenCalledTimes(1);
  });
});

describe("state.timeToRemainingSeconds()", () => {
  it("should return the number of seconds until the timestamp is reached", async () => {
    const state = class extends State {};
    const time = 1664368504;

    expect(state.timeToRemainingSeconds(time)).toEqual(
      time - Math.floor(Date.now() / 1000)
    );
  });
});

describe("state.remainingSecondsToTime()", () => {
  it("should return the timestamp when adding seconds to now", async () => {
    const state = class extends State {};
    const seconds = 42;

    expect(state.remainingSecondsToTime(seconds)).toEqual(
      Math.floor(Date.now() / 1000) + seconds
    );
  });
});
