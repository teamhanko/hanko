import JSCookie from "js-cookie";
import { Cookie } from "../../src/lib/Cookie";
import { fakeTimerNow } from "../setup";
import { Response } from "../../src/lib/client/HttpClient";

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
        sameSite: "lax",
        domain: undefined,
      });
    });

    it("should set a new secure cookie", async () => {
      jest.spyOn(JSCookie, "set");
      cookie.setAuthCookie("test-token");

      expect(JSCookie.set).toHaveBeenCalledWith("hanko", "test-token", {
        secure: true,
        domain: undefined,
        sameSite: "lax",
      });
    });

    it("should set a new cookie with expiration", async () => {
      jest.spyOn(JSCookie, "set");
      const expires = new Date(fakeTimerNow + 60);
      cookie.setAuthCookie("test-token", { secure: true, expires });
      expect(JSCookie.set).toHaveBeenCalledWith("hanko", "test-token", {
        secure: true,
        sameSite: "lax",
        domain: undefined,
        expires,
      });
    });

    it("should set a new cookie with given SameSite value", async () => {
      jest.spyOn(JSCookie, "set");
      cookie.setAuthCookie("test-token", { sameSite: "strict" });
      expect(JSCookie.set).toHaveBeenCalledWith("hanko", "test-token", {
        secure: true,
        sameSite: "strict",
      });
    });

    it("should throw if not Secure and SameSite value is none", async () => {
      jest.spyOn(JSCookie, "set");
      expect(() => {
        cookie.setAuthCookie("test-token", { secure: false, sameSite: "none" });
      }).toThrow("Technical error");
    });

    it("should set a new cookie with given domain value", async () => {
      jest.spyOn(JSCookie, "set");
      cookie.setAuthCookie("test-token", { domain: ".test.app" });
      expect(JSCookie.set).toHaveBeenCalledWith("hanko", "test-token", {
        secure: true,
        sameSite: "lax",
        domain: ".test.app",
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
