import {
  PasswordClient,
  InvalidPasswordError,
  TooManyRequestsError,
  TechnicalError,
} from "../../../src";
import { Response } from "../../../src/lib/client/HttpClient";

const userID = "test-user-1";
const password = "test-password-1";
const passwordRetryAfter = 180;
let passwordClient: PasswordClient;

beforeEach(() => {
  passwordClient = new PasswordClient("http://test.api");
});

describe("PasswordClient.login()", () => {
  it("should do a password login", async () => {
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
    response.ok = true;
    jest.spyOn(passwordClient.client, "post").mockResolvedValue(response);
    jest.spyOn(passwordClient.client, "processResponseHeadersOnLogin");

    const loginResponse = passwordClient.login(userID, password);
    await expect(loginResponse).resolves.toBeUndefined();
    expect(passwordClient.client.processResponseHeadersOnLogin).toHaveBeenCalledTimes(1);
    expect(passwordClient.client.post).toHaveBeenCalledWith("/password/login", {
      user_id: userID,
      password,
    });
  });

  it("should throw error when using an invalid password", async () => {
    const response = new Response(new XMLHttpRequest());
    response.status = 401;
    jest.spyOn(passwordClient.client, "post").mockResolvedValue(response);

    const loginResponse = passwordClient.login(userID, password);
    await expect(loginResponse).rejects.toThrow(InvalidPasswordError);
  });

  it("should throw error and set retry after in state on too many request response from API", async () => {
    const xhr = new XMLHttpRequest();
    const response = new Response(xhr);

    response.status = 429;

    jest.spyOn(passwordClient.client, "post").mockResolvedValue(response);
    jest
      .spyOn(response.headers, "getResponseHeader")
      .mockReturnValue(`${passwordRetryAfter}`);
    jest.spyOn(passwordClient.passwordState, "read");
    jest.spyOn(passwordClient.passwordState, "setRetryAfter");
    jest.spyOn(passwordClient.passwordState, "write");

    await expect(passwordClient.login(userID, password)).rejects.toThrowError(
      TooManyRequestsError
    );

    expect(passwordClient.passwordState.read).toHaveBeenCalledTimes(1);
    expect(passwordClient.passwordState.setRetryAfter).toHaveBeenCalledWith(
      userID,
      passwordRetryAfter
    );
    expect(passwordClient.passwordState.write).toHaveBeenCalledTimes(1);
    expect(response.headers.getResponseHeader).toHaveBeenCalledWith("Retry-After");
  });

  it("should throw error when API response is not ok", async () => {
    const response = new Response(new XMLHttpRequest());
    passwordClient.client.post = jest.fn().mockResolvedValue(response);

    const loginResponse = passwordClient.login(userID, password);
    await expect(loginResponse).rejects.toThrowError(TechnicalError);
  });

  it("should throw error on API communication failure", async () => {
    passwordClient.client.post = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const loginResponse = passwordClient.login(userID, password);
    await expect(loginResponse).rejects.toThrowError("Test error");
  });
});

describe("PasswordClient.update()", () => {
  it("should update a password", async () => {
    const response = new Response(new XMLHttpRequest());
    response.ok = true;
    jest.spyOn(passwordClient.client, "put").mockResolvedValue(response);

    const loginResponse = passwordClient.update(userID, password);
    await expect(loginResponse).resolves.toBeUndefined();

    expect(passwordClient.client.put).toHaveBeenCalledWith("/password", {
      user_id: userID,
      password,
    });
  });

  it.each`
    status | error
    ${401} | ${"Unauthorized error"}
    ${500} | ${"Technical error"}
  `(
    "should throw error when API response is not ok",
    async ({ status, error }) => {
      const response = new Response(new XMLHttpRequest());
      response.ok = status >= 200 && status <= 299;
      response.status = status;
      passwordClient.client.put = jest.fn().mockResolvedValue(response);

      const config = passwordClient.update(userID, password);
      await expect(config).rejects.toThrowError(error);
    }
  );

  it("should throw error on API communication failure", async () => {
    passwordClient.client.put = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const config = passwordClient.update(userID, password);
    await expect(config).rejects.toThrowError("Test error");
  });

  describe("PasswordClient.getRetryAfter()", () => {
    it("should return password resend after seconds", async () => {
      jest
        .spyOn(passwordClient.passwordState, "getRetryAfter")
        .mockReturnValue(passwordRetryAfter);
      expect(passwordClient.getRetryAfter(userID)).toEqual(passwordRetryAfter);
    });
  });
});
