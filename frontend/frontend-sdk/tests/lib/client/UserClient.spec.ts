import { UserClient } from "../../../src";
import { Response } from "../../../src/lib/client/HttpClient";

let userClient: UserClient;

beforeEach(() => {
  userClient = new UserClient("http://test.api", {
    cookieName: "hanko",
    timeout: 13000,
    sessionTokenLocation: "cookie",
  });
});

describe("UserClient.getCurrent()", () => {
  const userID = "test-user-1";
  const email = "test-email-1@test";
  const credentials = [{ id: "test-credential-1" }];

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
      `/users/${userID}`,
    );
  });

  it.each`
    statusMe | statusUsers | error
    ${400}   | ${200}      | ${"Technical error"}
    ${401}   | ${200}      | ${"Unauthorized error"}
    ${404}   | ${200}      | ${"Technical error"}
    ${200}   | ${400}      | ${"Technical error"}
    ${200}   | ${401}      | ${"Unauthorized error"}
    ${200}   | ${404}      | ${"Technical error"}
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
    },
  );

  it("should throw error on API communication failure", async () => {
    userClient.client.get = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const user = userClient.getCurrent();
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
    },
  );
});
