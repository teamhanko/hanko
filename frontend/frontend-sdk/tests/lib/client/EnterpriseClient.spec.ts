import {
  EnterpriseClient,
  NotFoundError,
  TechnicalError,
  ThirdPartyError,
} from "../../../src";
import { Response } from "../../../src/lib/client/HttpClient";

let enterpriseClient: EnterpriseClient;

beforeEach(() => {
  enterpriseClient = new EnterpriseClient("http://test.api", {
    cookieName: "hanko",
    localStorageKey: "hanko",
    timeout: 13000,
  });
});

describe("enterpriseClient.hasProvider()", () => {
  let response: Response;
  beforeEach(() => {
    response = new Response(new XMLHttpRequest());
    response.ok = true;
    response.status = 200;

    jest.spyOn(enterpriseClient.client, "get").mockResolvedValue(response);
  });

  it("should fetch provider with correct domain", async () => {
    const result = await enterpriseClient.hasProvider("test@test.example");
    const expectedUrl = "/saml/provider?domain=test.example";

    expect(enterpriseClient.client.get).toHaveBeenCalledWith(expectedUrl);
    expect(result).toBeTruthy();
  });

  it("should fail to fetch provider", async () => {
    response.ok = false;
    response.status = 404;
    jest.spyOn(enterpriseClient.client, "get").mockResolvedValue(response);

    await expect(
      enterpriseClient.hasProvider("test@nottest.example"),
    ).rejects.toThrow(NotFoundError);
  });

  it("should fail to fetch provider due to server error", async () => {
    response.ok = false;
    response.status = 400;
    jest.spyOn(enterpriseClient.client, "get").mockResolvedValue(response);

    await expect(
      enterpriseClient.hasProvider("test@nottest.example"),
    ).rejects.toThrow(TechnicalError);
  });

  it("should throw if email is empty", async () => {
    await expect(enterpriseClient.hasProvider("")).rejects.toThrow(
      ThirdPartyError,
    );

    expect(enterpriseClient.client.get).not.toHaveBeenCalled();
  });

  it("should throw if email has no domain", async () => {
    await expect(enterpriseClient.hasProvider("test@")).rejects.toThrow(
      ThirdPartyError,
    );

    expect(enterpriseClient.client.get).not.toHaveBeenCalled();
  });

  it("should throw if email has wrong format", async () => {
    await expect(enterpriseClient.hasProvider("test")).rejects.toThrow(
      ThirdPartyError,
    );

    expect(enterpriseClient.client.get).not.toHaveBeenCalled();
  });
});

describe("enterpriseClient.auth()", () => {
  const realLocation = window.location;

  beforeEach(() => {
    delete window.location;
    // @ts-ignore
    window.location = { ...realLocation, assign: jest.fn() };
  });

  afterEach(() => {
    window.location = realLocation;
  });

  it("should throw if email is empty", () => {
    expect(() => enterpriseClient.auth("", "http://test.example")).toThrow(
      ThirdPartyError,
    );
    expect(window.location.assign).not.toHaveBeenCalled();
  });

  it("should throw if email is wrong format", () => {
    expect(() => enterpriseClient.auth("test", "http://test.example")).toThrow(
      ThirdPartyError,
    );
    expect(window.location.assign).not.toHaveBeenCalled();
  });

  it("should throw if email is missing domain", () => {
    expect(() => enterpriseClient.auth("test@", "http://test.example")).toThrow(
      ThirdPartyError,
    );
    expect(window.location.assign).not.toHaveBeenCalled();
  });

  it("should throw if redirectTo is empty", () => {
    expect(() => enterpriseClient.auth("test@test.example", "")).toThrow(
      ThirdPartyError,
    );
    expect(window.location.assign).not.toHaveBeenCalled();
  });

  it("should construct correct redirect url with provider", () => {
    expect(() =>
      enterpriseClient.auth("test@test.example", "http://test.example"),
    ).not.toThrow();
    const expectedUrl =
      "http://test.api/saml/auth?domain=test.example&redirect_to=http%3A%2F%2Ftest.example";
    expect(window.location.assign).toHaveBeenCalledWith(expectedUrl);
  });
});

describe("enterpriseClient.getError()", () => {
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
    ${"access_denied"}     | ${"enterpriseAccessDenied"}
    ${"user_conflict"}     | ${"emailAddressAlreadyExistsError"}
    ${"multiple_accounts"} | ${"enterpriseMultipleAccounts"}
    ${"unverified_email"}  | ${"enterpriseUnverifiedEmail"}
    ${"email_maxnum"}      | ${"maxNumOfEmailAddressesReached"}
  `("should map to correct error", async ({ error, expectedCode }) => {
    window.location.search = `?error=${error}`;
    const got = enterpriseClient.getError();
    expect(got.code).toBe(expectedCode);
  });
});
