import { TechnicalError, TokenClient } from "../../../src";
import { Response } from "../../../src/lib/client/HttpClient";

let tokenClient: TokenClient;

beforeEach(() => {
  tokenClient = new TokenClient("http://test.api");
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
