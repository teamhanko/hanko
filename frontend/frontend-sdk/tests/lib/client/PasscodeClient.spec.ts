import {
  InvalidPasscodeError,
  MaxNumOfPasscodeAttemptsReachedError,
  PasscodeClient,
  PasscodeExpiredError,
  TechnicalError,
  TooManyRequestsError,
} from "../../../src";
import { Response } from "../../../src/lib/client/HttpClient";

const userID = "test-user-1";
const passcodeID = "test-passcode-1";
const emailID = "test-email-1";
const passcodeTTL = 180;
const passcodeRetryAfter = 180;
const passcodeValue = "123456";
let passcodeClient: PasscodeClient;

beforeEach(() => {
  passcodeClient = new PasscodeClient("http://test.api");
});

describe("PasscodeClient.initialize()", () => {
  it("should initialize a passcode login", async () => {
    const response = new Response(new XMLHttpRequest());
    response.ok = true;
    response._decodedJSON = { id: passcodeID, ttl: passcodeTTL };
    jest.spyOn(passcodeClient.client, "post").mockResolvedValue(response);

    jest.spyOn(passcodeClient.state, "read");
    jest.spyOn(passcodeClient.state, "setTTL");
    jest.spyOn(passcodeClient.state, "setActiveID");
    jest.spyOn(passcodeClient.state, "write");

    const passcode = await passcodeClient.initialize(userID);
    expect(passcode.id).toEqual(passcodeID);
    expect(passcode.ttl).toEqual(passcodeTTL);

    expect(passcodeClient.state.read).toHaveBeenCalledTimes(1);
    expect(passcodeClient.state.setTTL).toHaveBeenCalledWith(
      userID,
      passcodeTTL
    );
    expect(passcodeClient.state.setActiveID).toHaveBeenCalledWith(
      userID,
      passcodeID
    );
    expect(passcodeClient.state.write).toHaveBeenCalledTimes(1);
    expect(passcodeClient.client.post).toHaveBeenCalledWith(
      "/passcode/login/initialize",
      { user_id: userID }
    );
  });

  it("should initialize a passcode with specified email id", async () => {
    const response = new Response(new XMLHttpRequest());
    response.ok = true;

    jest.spyOn(passcodeClient.client, "post").mockResolvedValue(response);
    jest.spyOn(passcodeClient.state, "setEmailID");

    await passcodeClient.initialize(userID, emailID, true);

    expect(passcodeClient.state.setEmailID).toHaveBeenCalledWith(
      userID,
      emailID
    );
    expect(passcodeClient.client.post).toHaveBeenCalledWith(
      "/passcode/login/initialize",
      { user_id: userID, email_id: emailID }
    );
  });

  it("should restore the previous passcode", async () => {
    jest.spyOn(passcodeClient.state, "read");
    jest.spyOn(passcodeClient.state, "getTTL").mockReturnValue(passcodeTTL);
    jest.spyOn(passcodeClient.state, "getActiveID").mockReturnValue(passcodeID);
    jest.spyOn(passcodeClient.state, "getEmailID").mockReturnValue(emailID);

    await expect(passcodeClient.initialize(userID, emailID)).resolves.toEqual({
      id: passcodeID,
      ttl: passcodeTTL,
    });

    expect(passcodeClient.state.read).toHaveBeenCalledTimes(1);
    expect(passcodeClient.state.getTTL).toHaveBeenCalledWith(userID);
    expect(passcodeClient.state.getActiveID).toHaveBeenCalledWith(userID);
    expect(passcodeClient.state.getEmailID).toHaveBeenCalledWith(userID);
  });

  it("should throw an error as long as email backoff is active", async () => {
    jest
      .spyOn(passcodeClient.state, "getResendAfter")
      .mockReturnValue(passcodeRetryAfter);

    await expect(passcodeClient.initialize(userID, emailID)).rejects.toThrow(
      TooManyRequestsError
    );

    expect(passcodeClient.state.getResendAfter).toHaveBeenCalledWith(userID);
  });

  it("should throw error and set retry after in state on too many request response from API", async () => {
    const xhr = new XMLHttpRequest();
    const response = new Response(xhr);

    response.status = 429;

    jest.spyOn(passcodeClient.client, "post").mockResolvedValue(response);
    jest
      .spyOn(response.headers, "getResponseHeader")
      .mockReturnValue(`${passcodeRetryAfter}`);
    jest.spyOn(passcodeClient.state, "read");
    jest.spyOn(passcodeClient.state, "setResendAfter");
    jest.spyOn(passcodeClient.state, "write");

    await expect(passcodeClient.initialize(userID)).rejects.toThrowError(
      TooManyRequestsError
    );

    expect(passcodeClient.state.read).toHaveBeenCalledTimes(1);
    expect(passcodeClient.state.setResendAfter).toHaveBeenCalledWith(
      userID,
      passcodeRetryAfter
    );
    expect(passcodeClient.state.write).toHaveBeenCalledTimes(1);
    expect(response.headers.getResponseHeader).toHaveBeenCalledWith(
      "Retry-After"
    );
  });

  it.each`
    status | error
    ${401} | ${"Technical error"}
    ${500} | ${"Technical error"}
    ${429} | ${"Too many requests error"}
  `(
    "should throw error when API response is not ok",
    async ({ status, error }) => {
      const response = new Response(new XMLHttpRequest());
      response.status = status;
      response.ok = status >= 200 && status <= 299;

      passcodeClient.client.post = jest.fn().mockResolvedValue(response);

      const passcode = passcodeClient.initialize("test-user-1");
      await expect(passcode).rejects.toThrowError(error);
    }
  );

  it("should throw error on API communication failure", async () => {
    passcodeClient.client.post = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const passcode = passcodeClient.initialize("test-user-1");
    await expect(passcode).rejects.toThrowError("Test error");
  });
});

describe("PasscodeClient.finalize()", () => {
  it("should finalize a passcode login", async () => {
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

    jest.spyOn(passcodeClient.state, "read");
    jest.spyOn(passcodeClient.client, "processResponseHeadersOnLogin");
    jest.spyOn(passcodeClient.state, "getActiveID").mockReturnValue(passcodeID);
    jest.spyOn(passcodeClient.state, "getTTL").mockReturnValue(passcodeTTL);
    jest.spyOn(passcodeClient.client, "post").mockResolvedValue(response);

    await expect(
      passcodeClient.finalize(userID, passcodeValue)
    ).resolves.toBeUndefined();
    expect(passcodeClient.state.read).toHaveBeenCalledTimes(1);
    expect(
      passcodeClient.client.processResponseHeadersOnLogin
    ).toHaveBeenCalledTimes(1);
    expect(passcodeClient.state.getActiveID).toHaveBeenCalledWith(userID);
    expect(passcodeClient.client.post).toHaveBeenCalledWith(
      "/passcode/login/finalize",
      { id: passcodeID, code: passcodeValue }
    );
  });

  it("should throw error when using an invalid passcode", async () => {
    const response = new Response(new XMLHttpRequest());
    response.status = 401;

    jest.spyOn(passcodeClient.state, "read");
    jest.spyOn(passcodeClient.state, "getActiveID").mockReturnValue(passcodeID);
    jest.spyOn(passcodeClient.state, "getTTL").mockReturnValue(passcodeTTL);
    jest.spyOn(passcodeClient.client, "post").mockResolvedValue(response);

    await expect(
      passcodeClient.finalize(userID, passcodeValue)
    ).rejects.toThrow(InvalidPasscodeError);
    expect(passcodeClient.state.read).toHaveBeenCalledTimes(1);
    expect(passcodeClient.state.getActiveID).toHaveBeenCalledWith(userID);
  });

  it("should throw error when reaching max passcode attempts", async () => {
    const response = new Response(new XMLHttpRequest());
    response.status = 410;

    jest.spyOn(passcodeClient.state, "read");
    jest.spyOn(passcodeClient.state, "reset");
    jest.spyOn(passcodeClient.state, "write");
    jest.spyOn(passcodeClient.state, "getActiveID").mockReturnValue(passcodeID);
    jest.spyOn(passcodeClient.state, "getTTL").mockReturnValue(passcodeTTL);
    jest.spyOn(passcodeClient.client, "post").mockResolvedValue(response);

    await expect(
      passcodeClient.finalize(userID, passcodeValue)
    ).rejects.toThrow(MaxNumOfPasscodeAttemptsReachedError);
    expect(passcodeClient.state.read).toHaveBeenCalledTimes(1);
    expect(passcodeClient.state.reset).toHaveBeenCalledTimes(1);
    expect(passcodeClient.state.write).toHaveBeenCalledTimes(1);
    expect(passcodeClient.state.getActiveID).toHaveBeenCalledWith(userID);
  });

  it("should throw error when the passcode has expired", async () => {
    jest.spyOn(passcodeClient.state, "getTTL").mockReturnValue(0);
    const finalizeResponse = passcodeClient.finalize(userID, passcodeValue);
    await expect(finalizeResponse).rejects.toThrowError(PasscodeExpiredError);
  });

  it("should throw error when API response is not ok", async () => {
    const response = new Response(new XMLHttpRequest());
    passcodeClient.client.post = jest.fn().mockResolvedValue(response);
    jest.spyOn(passcodeClient.state, "getTTL").mockReturnValue(passcodeTTL);

    const finalizeResponse = passcodeClient.finalize(userID, passcodeValue);
    await expect(finalizeResponse).rejects.toThrowError(TechnicalError);
  });

  it("should throw error on API communication failure", async () => {
    passcodeClient.client.post = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));
    jest.spyOn(passcodeClient.state, "getTTL").mockReturnValue(passcodeTTL);

    const finalizeResponse = passcodeClient.finalize(userID, passcodeValue);
    await expect(finalizeResponse).rejects.toThrowError("Test error");
  });
});

describe("PasscodeClient.getTTL()", () => {
  it("should return passcode TTL", async () => {
    jest.spyOn(passcodeClient.state, "getTTL").mockReturnValue(passcodeTTL);
    expect(passcodeClient.getTTL(userID)).toEqual(passcodeTTL);
  });
});

describe("PasscodeClient.getResendAfter()", () => {
  it("should return passcode resend after seconds", async () => {
    jest
      .spyOn(passcodeClient.state, "getResendAfter")
      .mockReturnValue(passcodeRetryAfter);
    expect(passcodeClient.getResendAfter(userID)).toEqual(passcodeRetryAfter);
  });
});
