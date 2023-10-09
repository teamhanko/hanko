import JSCookie from "js-cookie";
import { Cookie } from "../../src/lib/Cookie";
import { fakeTimerNow } from "../setup";

describe("Cookie()", () => {
  let cookie: Cookie;

  beforeEach(() => {
    cookie = new Cookie({ cookieName: "hanko" });
  });

  describe("cookie.setAuthCookie()", () => {
    it("should set a new cookie", async () => {
      jest.spyOn(JSCookie, "set");
      cookie.setAuthCookie("test-token", { secure: false });
      expect(JSCookie.set).toHaveBeenCalledWith("hanko", "test-token", {
        secure: false,
      });
    });

    it("should set a new secure cookie", async () => {
      jest.spyOn(JSCookie, "set");
      cookie.setAuthCookie("test-token");

      expect(JSCookie.set).toHaveBeenCalledWith("hanko", "test-token", {
        secure: true,
      });
    });

    it("should set a new cookie with expiration", async () => {
      jest.spyOn(JSCookie, "set");
      const expires = new Date(fakeTimerNow + 60);
      cookie.setAuthCookie("test-token", { secure: true, expires });
      expect(JSCookie.set).toHaveBeenCalledWith("hanko", "test-token", {
        secure: true,
        expires,
      });
    });
  });

  describe("cookie.getAuthCookie()", () => {
    it("should return the contents of the authorization cookie", async () => {
      JSCookie.get = jest.fn().mockReturnValue("test-token");
      const token = cookie.getAuthCookie();

      expect(JSCookie.get).toHaveBeenCalledWith("hanko");
      expect(token).toBe("test-token");
    });
  });

  describe("cookie.removeAuthCookie()", () => {
    it("should return the contents of the authorization cookie", async () => {
      jest.spyOn(JSCookie, "remove");
      cookie.removeAuthCookie();

      expect(JSCookie.remove).toHaveBeenCalledWith("hanko");
    });
  });
});
