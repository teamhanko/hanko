import {
  ConflictError,
  NotFoundError,
  TechnicalError,
  UserClient,
} from "../../../src";
import { Response } from "../../../src/lib/client/HttpClient";

const userID = "test-user-1";
const emailID = "test-email-1";
const email = "test-email-1@test";
const credentials = [{ id: "test-credential-1" }];

let userClient: UserClient;

beforeEach(() => {
  userClient = new UserClient("http://test.api");
});

describe("UserClient.getInfo()", () => {
  it("should retrieve user info", async () => {
    const response = new Response(new XMLHttpRequest());
    response.ok = true;
    response._decodedJSON = {
      id: userID,
      verified: true,
      has_webauthn_credential: true,
    };

    jest.spyOn(userClient.client, "post").mockResolvedValueOnce(response);
    const getInfoResponse = userClient.getInfo(email);
    await expect(getInfoResponse).resolves.toBe(response._decodedJSON);

    expect(userClient.client.post).toHaveBeenCalledWith("/user", {
      email,
    });
  });

  it("should throw error when user not found", async () => {
    const response = new Response(new XMLHttpRequest());
    response.status = 404;
    jest.spyOn(userClient.client, "post").mockResolvedValue(response);

    const user = userClient.getInfo(email);
    await expect(user).rejects.toThrow(NotFoundError);
  });

  it("should throw error when API response is not ok", async () => {
    const response = new Response(new XMLHttpRequest());
    userClient.client.post = jest.fn().mockResolvedValue(response);

    const user = userClient.getInfo(email);
    await expect(user).rejects.toThrowError(TechnicalError);
  });

  it("should throw error on API communication failure", async () => {
    userClient.client.post = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const user = userClient.getInfo(email);
    await expect(user).rejects.toThrowError("Test error");
  });
});

describe("UserClient.getCurrent()", () => {
  it("should retrieve currently logged in user", async () => {
    const responseMe = new Response(new XMLHttpRequest());
    responseMe.ok = true;
    responseMe._decodedJSON = {
      id: userID,
    };

    const responseUser = new Response(new XMLHttpRequest());
    responseUser.ok = true;
    responseUser._decodedJSON = {
      id: userID,
      email,
      webauthn_credentials: credentials,
    };

    jest
      .spyOn(userClient.client, "get")
      .mockResolvedValueOnce(responseMe)
      .mockResolvedValueOnce(responseUser);

    const user = userClient.getCurrent();
    await expect(user).resolves.toBe(responseUser._decodedJSON);

    expect(userClient.client.get).toHaveBeenNthCalledWith(1, "/me");
    expect(userClient.client.get).toHaveBeenNthCalledWith(
      2,
      `/users/${userID}`
    );
  });

  it.each`
    statusMe | statusUsers | error
    ${400}   | ${200}      | ${"Unauthorized error"}
    ${401}   | ${200}      | ${"Unauthorized error"}
    ${404}   | ${200}      | ${"Unauthorized error"}
    ${200}   | ${400}      | ${"Unauthorized error"}
    ${200}   | ${401}      | ${"Unauthorized error"}
    ${200}   | ${404}      | ${"Unauthorized error"}
    ${200}   | ${500}      | ${"Technical error"}
    ${500}   | ${200}      | ${"Technical error"}
  `(
    "should throw error if API returns an error status",
    async ({ statusMe, statusUsers, error }) => {
      const responseMe = new Response(new XMLHttpRequest());
      responseMe.status = statusMe;
      responseMe.ok = statusMe >= 200 && statusMe <= 299;

      const responseUser = new Response(new XMLHttpRequest());
      responseUser.status = statusUsers;
      responseUser.ok = statusUsers >= 200 && statusUsers <= 299;

      jest
        .spyOn(userClient.client, "get")
        .mockResolvedValueOnce(responseMe)
        .mockResolvedValueOnce(responseUser);

      const user = userClient.getCurrent();
      await expect(user).rejects.toThrow(error);
    }
  );

  it("should throw error on API communication failure", async () => {
    userClient.client.get = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const user = userClient.getCurrent();
    await expect(user).rejects.toThrowError("Test error");
  });
});

describe("UserClient.create()", () => {
  it("should create a user", async () => {
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
    response._decodedJSON = {
      user_id: userID,
      email_id: emailID,
    };

    jest.spyOn(userClient.client, "post").mockResolvedValueOnce(response);
    const getInfoResponse = userClient.create(email);
    await expect(getInfoResponse).resolves.toBe(response._decodedJSON);

    expect(userClient.client.post).toHaveBeenCalledWith("/users", {
      email,
    });
  });

  it("should throw error when user already exists", async () => {
    const response = new Response(new XMLHttpRequest());
    response.status = 409;
    jest.spyOn(userClient.client, "post").mockResolvedValue(response);

    const user = userClient.create(email);
    await expect(user).rejects.toThrow(ConflictError);
  });

  it("should throw error if API response is not ok (no 2xx, no 4xx)", async () => {
    const response = new Response(new XMLHttpRequest());
    jest.spyOn(userClient.client, "post").mockResolvedValue(response);

    const user = userClient.create(email);
    await expect(user).rejects.toThrow(TechnicalError);
  });

  it("should throw error on API communication failure", async () => {
    userClient.client.post = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const user = userClient.create(email);
    await expect(user).rejects.toThrowError("Test error");
  });
});

describe("UserClient.logout()", () => {
  it.each`
    status
    ${200}
    ${401}
  `("should return true if logout is successful", async ({ status }) => {
    const response = new Response(new XMLHttpRequest());
    response.status = status;
    response.ok = status >= 200 && status <= 299;

    jest.spyOn(userClient.client, "post").mockResolvedValueOnce(response);
    await expect(userClient.logout()).resolves.not.toThrow();

    expect(userClient.client.post).toHaveBeenCalledWith("/logout");
  });

  it.each`
    status | error
    ${400} | ${"Technical error"}
    ${404} | ${"Technical error"}
    ${500} | ${"Technical error"}
  `(
    "should throw error if API returns an error status",
    async ({ status, error }) => {
      const response = new Response(new XMLHttpRequest());
      response.status = status;
      response.ok = status >= 200 && status <= 299;

      jest.spyOn(userClient.client, "post").mockResolvedValueOnce(response);

      await expect(userClient.logout()).rejects.toThrow(error);

      expect(userClient.client.post).toHaveBeenCalledWith("/logout");
    }
  );
});

describe("UserClient.delete()", () => {
  it("should return true if deletion is successful", async () => {
    const response = new Response(new XMLHttpRequest());
    response.status = 204;
    response.ok = true;

    jest.spyOn(userClient.client, "delete").mockResolvedValueOnce(response);

    await expect(userClient.delete()).resolves.not.toThrow();
    expect(userClient.client.delete).toHaveBeenCalledWith("/user");
  });

  it.each`
    status | error
    ${401} | ${"Unauthorized error"}
    ${404} | ${"Technical error"}
    ${500} | ${"Technical error"}
  `(
    "should throw error if API returns an error status",
    async ({ status, error }) => {
      const response = new Response(new XMLHttpRequest());
      response.status = status;
      response.ok = status >= 200 && status <= 299;

      jest.spyOn(userClient.client, "delete").mockResolvedValueOnce(response);

      await expect(userClient.delete()).rejects.toThrow(error);

      expect(userClient.client.delete).toHaveBeenCalledWith("/user");
    }
  );
});
