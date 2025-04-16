import { UserClient } from "../../../src";
import { Response } from "../../../src/lib/client/HttpClient";

let userClient: UserClient;

beforeEach(() => {
  userClient = new UserClient("http://test.api", {
    cookieName: "hanko",
    timeout: 13000,
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
