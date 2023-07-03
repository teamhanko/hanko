import { ThirdPartyClient } from "../../../src";
import { ThirdPartyError } from "../../../src/lib/Errors";

let thirdPartyClient: ThirdPartyClient;

beforeEach(() => {
  thirdPartyClient = new ThirdPartyClient("http://test.api", {});
});

describe("thirdPartyClient.auth()", () => {
  const realLocation = window.location;

  beforeEach(() => {
    delete window.location;
    // @ts-ignore
    window.location = { ...realLocation, assign: jest.fn() };
  });

  afterEach(() => {
    window.location = realLocation;
  });

  it("should throw if provider is empty", async () => {
    await expect(
      thirdPartyClient.auth("", "http://test.example")
    ).rejects.toThrow(ThirdPartyError);
    expect(window.location.assign).not.toHaveBeenCalled();
  });

  it("should throw if redirectTo is empty", async () => {
    await expect(thirdPartyClient.auth("testProvider", "")).rejects.toThrow(
      ThirdPartyError
    );
    expect(window.location.assign).not.toHaveBeenCalled();
  });

  it("should construct correct redirect url with provider", async () => {
    await expect(
      thirdPartyClient.auth("testProvider", "http://test.example")
    ).resolves.not.toThrow();
    const expectedUrl =
      "http://test.api/thirdparty/auth?provider=testProvider&redirect_to=http%3A%2F%2Ftest.example";
    expect(window.location.assign).toHaveBeenCalledWith(expectedUrl);
  });
});

describe("thirdPartyClient.getError()", () => {
  const realLocation = window.location;

  beforeEach(() => {
    delete window.location;
    // @ts-ignore
    window.location = {
      search: "",
    };
  });

  afterEach(() => {
    window.location = realLocation;
  });

  it.each`
    error                  | expectedCode
    ${"server_error"}      | ${"somethingWentWrong"}
    ${"invalid_request"}   | ${"somethingWentWrong"}
    ${"access_denied"}     | ${"thirdPartyAccessDenied"}
    ${"user_conflict"}     | ${"emailAddressAlreadyExistsError"}
    ${"multiple_accounts"} | ${"thirdPartyMultipleAccounts"}
    ${"unverified_email"}  | ${"thirdPartyUnverifiedEmail"}
    ${"email_maxnum"}      | ${"maxNumOfEmailAddressesReached"}
  `("should map to correct error", async ({ error, expectedCode }) => {
    window.location.search = `?error=${error}`;
    const got = thirdPartyClient.getError();
    expect(got.code).toBe(expectedCode);
  });
});
