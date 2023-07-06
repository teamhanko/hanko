import { TechnicalError, TokenClient } from "../../../src";
import { Response } from "../../../src/lib/client/HttpClient";

let tokenClient: TokenClient;

beforeEach(() => {
  tokenClient = new TokenClient("http://test.api", {
    cookieName: "hanko",
    localStorageKey: "hanko",
    timeout: 13000,
  });
});

describe("tokenClient.validate()", () => {
  const realLocation = window.location;

  beforeEach(() => {
    delete window.location;
    // @ts-ignore
    window.location = {
      search: "",
      pathname: "",
    };
  });

  afterEach(() => {
    window.location = realLocation;
  });

  it("should resolve when no token in url params", async () => {
    await expect(tokenClient.validate()).resolves.not.toThrow();
  });

  it("should return technical error on API error response", async () => {
    window.location.search = "?hanko_token=invalid_token";

    const response = new Response(new XMLHttpRequest());
    response.status = 400;
    response.ok = false;

    jest.spyOn(tokenClient.client, "post").mockResolvedValue(response);

    await expect(tokenClient.validate()).rejects.toThrow(TechnicalError);
  });

  describe("tokenClient.validate() - on success", () => {
    beforeEach(() => {
      Object.defineProperty(window, "history", {
        value: {
          replaceState: jest.fn(),
        },
      });
    });

    it("should exchange a token for a JWT", async () => {
      window.location.search = "?hanko_token=valid_token";
      window.location.pathname = "/callback";

      Object.defineProperty(global, "XMLHttpRequest", {
        value: jest.fn().mockImplementation(() => ({
          response: JSON.stringify({ foo: "bar" }),
          open: jest.fn(),
          setRequestHeader: jest.fn(),
          getResponseHeader: jest.fn(),
          getAllResponseHeaders: jest.fn().mockReturnValue(""),
          send: jest.fn(),
        })),
        configurable: true,
        writable: true,
      });

      const response = new Response(new XMLHttpRequest());
      response.status = 200;
      response.ok = true;

      jest.spyOn(tokenClient.client, "post").mockResolvedValue(response);

      await expect(tokenClient.validate()).resolves.not.toThrow();
      expect(window.history.replaceState).toHaveBeenCalledWith(
        null,
        null,
        "/callback"
      );
    });
  });
});
