import JSCookie from "js-cookie";
import { Cookie } from "../../src/lib/Cookie";

describe("Cookie()", () => {
  let cookie: Cookie;

  beforeEach(() => {
    cookie = new Cookie("hanko");
  });

  describe("cookie.setAuthCookie()", () => {
    it("should set a new cookie", async () => {
      jest.spyOn(JSCookie, "set");
      cookie.setAuthCookie("test-token", false);
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
