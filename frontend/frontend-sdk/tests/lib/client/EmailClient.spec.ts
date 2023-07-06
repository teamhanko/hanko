import { EmailClient } from "../../../src";
import { Response } from "../../../src/lib/client/HttpClient";

const emailID = "test-email-1";
const emailAddress = "test-email-1@test";

let emailClient: EmailClient;

beforeEach(() => {
  emailClient = new EmailClient("http://test.api", {
    cookieName: "hanko",
    localStorageKey: "hanko",
    timeout: 13000,
  });
});

describe("EmailClient.list()", () => {
  it("should list email addresses", async () => {
    const response = new Response(new XMLHttpRequest());
    response.ok = true;
    response._decodedJSON = [
      {
        id: emailID,
        address: emailAddress,
        is_verified: false,
        is_primary: true,
      },
    ];

    jest.spyOn(emailClient.client, "get").mockResolvedValue(response);
    const list = await emailClient.list();
    expect(emailClient.client.get).toHaveBeenCalledWith("/emails");
    expect(list).toEqual(response._decodedJSON);
  });
  it.each`
    status | error
    ${401} | ${"Unauthorized error"}
    ${500} | ${"Technical error"}
  `(
    "should throw error if API returns an error status",
    async ({ status, error }) => {
      const response = new Response(new XMLHttpRequest());
      response.status = status;
      response.ok = status >= 200 && status <= 299;

      jest.spyOn(emailClient.client, "get").mockResolvedValueOnce(response);

      const email = emailClient.list();
      await expect(email).rejects.toThrow(error);
    }
  );

  it("should throw error on API communication failure", async () => {
    emailClient.client.get = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const user = emailClient.list();
    await expect(user).rejects.toThrowError("Test error");
  });
});

describe("EmailClient.create()", () => {
  it("should create a email address", async () => {
    const response = new Response(new XMLHttpRequest());
    response.ok = true;
    response._decodedJSON = {
      id: "",
      address: "",
      is_verified: false,
      is_primary: true,
    };

    jest.spyOn(emailClient.client, "post").mockResolvedValue(response);

    const createResponse = emailClient.create(emailAddress);
    await expect(createResponse).resolves.toBe(response._decodedJSON);

    expect(emailClient.client.post).toHaveBeenCalledWith(`/emails`, {
      address: emailAddress,
    });
  });

  it.each`
    status | error
    ${400} | ${"The email address already exists"}
    ${401} | ${"Unauthorized error"}
    ${409} | ${"Maximum number of email addresses reached error"}
    ${500} | ${"Technical error"}
  `(
    "should throw error if API returns an error status",
    async ({ status, error }) => {
      const response = new Response(new XMLHttpRequest());
      response.status = status;
      response.ok = status >= 200 && status <= 299;

      jest.spyOn(emailClient.client, "post").mockResolvedValueOnce(response);

      const email = emailClient.create(emailAddress);
      await expect(email).rejects.toThrow(error);
    }
  );

  it("should throw error on API communication failure", async () => {
    emailClient.client.post = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const user = emailClient.create(emailAddress);
    await expect(user).rejects.toThrowError("Test error");
  });
});

describe("EmailClient.setPrimaryEmail()", () => {
  it("should set a primary email address", async () => {
    const response = new Response(new XMLHttpRequest());
    response.ok = true;

    jest.spyOn(emailClient.client, "post").mockResolvedValue(response);
    const update = await emailClient.setPrimaryEmail(emailID);
    expect(emailClient.client.post).toHaveBeenCalledWith(
      `/emails/${emailID}/set_primary`
    );
    expect(update).toEqual(undefined);
  });

  it.each`
    status | error
    ${401} | ${"Unauthorized error"}
    ${500} | ${"Technical error"}
  `(
    "should throw error if API returns an error status",
    async ({ status, error }) => {
      const response = new Response(new XMLHttpRequest());
      response.status = status;
      response.ok = status >= 200 && status <= 299;

      jest.spyOn(emailClient.client, "post").mockResolvedValueOnce(response);

      const email = emailClient.setPrimaryEmail(emailID);
      await expect(email).rejects.toThrow(error);
    }
  );

  it("should throw error on API communication failure", async () => {
    emailClient.client.post = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const user = emailClient.setPrimaryEmail(emailID);
    await expect(user).rejects.toThrowError("Test error");
  });
});

describe("EmailClient.delete()", () => {
  it("should delete email addresses", async () => {
    const response = new Response(new XMLHttpRequest());
    response.ok = true;

    jest.spyOn(emailClient.client, "delete").mockResolvedValue(response);
    const deleteResponse = await emailClient.delete(emailID);
    expect(emailClient.client.delete).toHaveBeenCalledWith(
      `/emails/${emailID}`
    );
    expect(deleteResponse).toEqual(undefined);
  });

  it.each`
    status | error
    ${401} | ${"Unauthorized error"}
    ${500} | ${"Technical error"}
  `(
    "should throw error if API returns an error status",
    async ({ status, error }) => {
      const response = new Response(new XMLHttpRequest());
      response.status = status;
      response.ok = status >= 200 && status <= 299;

      jest.spyOn(emailClient.client, "delete").mockResolvedValueOnce(response);

      const deleteResponse = emailClient.delete(emailID);
      await expect(deleteResponse).rejects.toThrow(error);
    }
  );

  it("should throw error on API communication failure", async () => {
    emailClient.client.delete = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const user = emailClient.delete(emailID);
    await expect(user).rejects.toThrowError("Test error");
  });
});
