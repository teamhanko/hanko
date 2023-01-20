import {
  InvalidPasscodeError,
  MaxNumOfPasscodeAttemptsReachedError,
  PasscodeClient,
  TechnicalError,
  TooManyRequestsError,
} from "../../../src";
import { Response } from "../../../src/lib/client/HttpClient";

const userID = "test-user-1";
const passcodeID = "test-passcode-1";
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

  it("should throw error and set retry after in state on too many request response from API", async () => {
    const xhr = new XMLHttpRequest();
    const response = new Response(xhr);

    response.status = 429;

    jest.spyOn(passcodeClient.client, "post").mockResolvedValue(response);
    jest
      .spyOn(response.headers, "get")
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
    expect(response.headers.get).toHaveBeenCalledWith("Retry-After");
  });

  it("should throw error when API response is not ok", async () => {
    const response = new Response(new XMLHttpRequest());
    passcodeClient.client.post = jest.fn().mockResolvedValue(response);

    const config = passcodeClient.initialize("test-user-1");
    await expect(config).rejects.toThrowError(TechnicalError);
  });

  it("should throw error on API communication failure", async () => {
    passcodeClient.client.post = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const config = passcodeClient.initialize("test-user-1");
    await expect(config).rejects.toThrowError("Test error");
  });
});

describe("PasscodeClient.finalize()", () => {
  it("should finalize a passcode login", async () => {
    const response = new Response(new XMLHttpRequest());
    response.ok = true;

    jest.spyOn(passcodeClient.state, "read");
    jest.spyOn(passcodeClient.state, "reset");
    jest.spyOn(passcodeClient.state, "write");
    jest.spyOn(passcodeClient.state, "getActiveID").mockReturnValue(passcodeID);
    jest.spyOn(passcodeClient.client, "post").mockResolvedValue(response);

    await expect(
      passcodeClient.finalize(userID, passcodeValue)
    ).resolves.toBeUndefined();
    expect(passcodeClient.state.read).toHaveBeenCalledTimes(1);
    expect(passcodeClient.state.reset).toHaveBeenCalledTimes(1);
    expect(passcodeClient.state.write).toHaveBeenCalledTimes(1);
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
    jest.spyOn(passcodeClient.client, "post").mockResolvedValue(response);

    await expect(
      passcodeClient.finalize(userID, passcodeValue)
    ).rejects.toThrow(MaxNumOfPasscodeAttemptsReachedError);
    expect(passcodeClient.state.read).toHaveBeenCalledTimes(1);
    expect(passcodeClient.state.reset).toHaveBeenCalledTimes(1);
    expect(passcodeClient.state.write).toHaveBeenCalledTimes(1);
    expect(passcodeClient.state.getActiveID).toHaveBeenCalledWith(userID);
  });

  it("should throw error when API response is not ok", async () => {
    const response = new Response(new XMLHttpRequest());
    passcodeClient.client.post = jest.fn().mockResolvedValue(response);

    const finalizeResponse = passcodeClient.finalize(userID, passcodeValue);
    await expect(finalizeResponse).rejects.toThrowError(TechnicalError);
  });

  it("should throw error on API communication failure", async () => {
    passcodeClient.client.post = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

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
