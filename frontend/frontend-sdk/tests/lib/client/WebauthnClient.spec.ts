import {
  User,
  WebauthnClient,
  WebauthnRequestCancelledError,
  WebauthnSupport,
} from "../../../src";
import { Response } from "../../../src/lib/client/HttpClient";

import { PublicKeyCredentialWithAssertionJSON } from "@github/webauthn-json";
import { Attestation } from "../../../src/lib/Dto";

const userID = "test-user-1";
const credentialID = "credential-1";

let webauthnClient: WebauthnClient;

beforeEach(() => {
  webauthnClient = new WebauthnClient("http://test.api");
});

describe("webauthnClient.login()", () => {
  const fakeRequestOptions = {} as PublicKeyCredentialRequestOptions;
  const fakeAssertion = {} as PublicKeyCredentialWithAssertionJSON;

  it("should perform a webauthn login", async () => {
    const initResponse = new Response(new XMLHttpRequest());
    initResponse.ok = true;
    initResponse._decodedJSON = fakeRequestOptions;

    const finalResponse = new Response(new XMLHttpRequest());
    finalResponse.ok = true;
    finalResponse._decodedJSON = {
      user_id: userID,
      credential_id: credentialID,
    };

    webauthnClient._getCredential = jest.fn().mockResolvedValue(fakeAssertion);
    webauthnClient._createAbortSignal = jest.fn();

    jest
      .spyOn(webauthnClient.client, "post")
      .mockResolvedValueOnce(initResponse)
      .mockResolvedValueOnce(finalResponse);

    jest.spyOn(webauthnClient.webauthnState, "read");
    jest.spyOn(webauthnClient.webauthnState, "addCredential");
    jest.spyOn(webauthnClient.webauthnState, "write");
    jest.spyOn(webauthnClient.passcodeState, "read");
    jest.spyOn(webauthnClient.passcodeState, "reset");
    jest.spyOn(webauthnClient.passcodeState, "write");

    await webauthnClient.login(userID, true);

    expect(webauthnClient._getCredential).toHaveBeenCalledWith({
      ...fakeRequestOptions,
      mediation: "conditional",
    });
    expect(webauthnClient._createAbortSignal).toHaveBeenCalledTimes(1);
    expect(webauthnClient.webauthnState.read).toHaveBeenCalledTimes(1);
    expect(webauthnClient.webauthnState.addCredential).toHaveBeenCalledWith(
      userID,
      credentialID
    );
    expect(webauthnClient.webauthnState.write).toHaveBeenCalledTimes(1);
    expect(webauthnClient.passcodeState.read).toHaveBeenCalledTimes(1);
    expect(webauthnClient.passcodeState.reset).toHaveBeenCalledWith(userID);
    expect(webauthnClient.passcodeState.write).toHaveBeenCalledTimes(1);
    expect(webauthnClient.client.post).toHaveBeenNthCalledWith(
      1,
      "/webauthn/login/initialize",
      { user_id: userID }
    );
    expect(webauthnClient.client.post).toHaveBeenNthCalledWith(
      2,
      "/webauthn/login/finalize",
      fakeAssertion
    );
  });

  it.each`
    statusInit | statusFinal | error
    ${500}     | ${200}      | ${"Technical error"}
    ${200}     | ${400}      | ${"Invalid WebAuthn credential error"}
    ${200}     | ${401}      | ${"Invalid WebAuthn credential error"}
    ${200}     | ${500}      | ${"Technical error"}
  `(
    "should throw error if API returns an error status",
    async ({ statusInit, statusFinal, error }) => {
      const initResponse = new Response(new XMLHttpRequest());
      initResponse.ok = statusInit >= 200 && statusInit <= 299;
      initResponse.status = statusInit;
      initResponse._decodedJSON = fakeRequestOptions;

      const finalResponse = new Response(new XMLHttpRequest());
      finalResponse.ok = statusFinal >= 200 && statusFinal <= 299;
      finalResponse.status = statusFinal;
      finalResponse._decodedJSON = {
        user_id: userID,
        credential_id: credentialID,
      };

      webauthnClient._getCredential = jest
        .fn()
        .mockResolvedValue({} as PublicKeyCredentialWithAssertionJSON);

      jest
        .spyOn(webauthnClient.client, "post")
        .mockResolvedValueOnce(initResponse)
        .mockResolvedValueOnce(finalResponse);

      const user = webauthnClient.login();
      await expect(user).rejects.toThrow(error);
    }
  );

  it("should throw an error when the WebAuthn API call fails", async () => {
    const initResponse = new Response(new XMLHttpRequest());
    initResponse.ok = true;

    jest
      .spyOn(webauthnClient.client, "post")
      .mockResolvedValueOnce(initResponse);

    webauthnClient._getCredential = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const user = webauthnClient.login();
    await expect(user).rejects.toThrow(WebauthnRequestCancelledError);
  });
});

describe("webauthnClient.register()", () => {
  const fakeCreationOptions = {} as PublicKeyCredentialCreationOptions;
  const fakeAttestation = { response: { transports: [] } } as Attestation;

  it("should perform a webauthn registration", async () => {
    const initResponse = new Response(new XMLHttpRequest());
    initResponse.ok = true;
    initResponse._decodedJSON = fakeCreationOptions;

    const finalResponse = new Response(new XMLHttpRequest());
    finalResponse.ok = true;
    finalResponse._decodedJSON = {
      user_id: userID,
      credential_id: credentialID,
    };

    webauthnClient._createCredential = jest
      .fn()
      .mockResolvedValue(fakeAttestation);
    webauthnClient._createAbortSignal = jest.fn();

    jest
      .spyOn(webauthnClient.client, "post")
      .mockResolvedValueOnce(initResponse)
      .mockResolvedValueOnce(finalResponse);

    jest.spyOn(webauthnClient.webauthnState, "read");
    jest.spyOn(webauthnClient.webauthnState, "addCredential");
    jest.spyOn(webauthnClient.webauthnState, "write");

    await webauthnClient.register();

    expect(webauthnClient._createCredential).toHaveBeenCalledWith({
      ...fakeCreationOptions,
    });
    expect(webauthnClient._createAbortSignal).toHaveBeenCalledTimes(1);
    expect(webauthnClient.webauthnState.read).toHaveBeenCalledTimes(1);
    expect(webauthnClient.webauthnState.addCredential).toHaveBeenCalledWith(
      userID,
      credentialID
    );
    expect(webauthnClient.webauthnState.write).toHaveBeenCalledTimes(1);
    expect(webauthnClient.client.post).toHaveBeenNthCalledWith(
      1,
      "/webauthn/registration/initialize"
    );
    expect(webauthnClient.client.post).toHaveBeenNthCalledWith(
      2,
      "/webauthn/registration/finalize",
      fakeAttestation
    );
  });

  it.each`
    statusInit | statusFinal | error
    ${400}     | ${200}      | ${"Unauthorized error"}
    ${500}     | ${200}      | ${"Technical error"}
    ${200}     | ${400}      | ${"Unauthorized error"}
    ${200}     | ${422}      | ${"User verification error"}
    ${200}     | ${500}      | ${"Technical error"}
  `(
    "should throw error if API returns an error status",
    async ({ statusInit, statusFinal, error }) => {
      const initResponse = new Response(new XMLHttpRequest());
      initResponse.ok = statusInit >= 200 && statusInit <= 299;
      initResponse.status = statusInit;
      initResponse._decodedJSON = fakeCreationOptions;

      const finalResponse = new Response(new XMLHttpRequest());
      finalResponse.ok = statusFinal >= 200 && statusFinal <= 299;
      finalResponse.status = statusFinal;
      finalResponse._decodedJSON = {
        user_id: userID,
        credential_id: credentialID,
      };

      webauthnClient._createCredential = jest
        .fn()
        .mockResolvedValue(fakeAttestation);

      jest
        .spyOn(webauthnClient.client, "post")
        .mockResolvedValueOnce(initResponse)
        .mockResolvedValueOnce(finalResponse);

      const user = webauthnClient.register();
      await expect(user).rejects.toThrow(error);
    }
  );

  it("should throw an error when the WebAuthn API call fails", async () => {
    const initResponse = new Response(new XMLHttpRequest());
    initResponse.ok = true;

    jest
      .spyOn(webauthnClient.client, "post")
      .mockResolvedValueOnce(initResponse);

    webauthnClient._createCredential = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const user = webauthnClient.register();
    await expect(user).rejects.toThrow(WebauthnRequestCancelledError);
  });
});

describe("webauthnClient.shouldRegister()", () => {
  it.each`
    isSupported | userHasCredential | credentialMatched | expected
    ${false}    | ${false}          | ${false}          | ${false}
    ${true}     | ${false}          | ${false}          | ${true}
    ${true}     | ${true}           | ${false}          | ${true}
    ${true}     | ${true}           | ${true}           | ${false}
  `(
    "should determine correctly if a WebAuthn credential should be registered",
    async ({ isSupported, userHasCredential, credentialMatched, expected }) => {
      jest.spyOn(WebauthnSupport, "supported").mockReturnValue(isSupported);

      const user: User = {
        id: userID,
        email_id: "",
        webauthn_credentials: [],
      };

      if (userHasCredential) {
        user.webauthn_credentials.push({ id: credentialID });
      }

      if (credentialMatched) {
        jest
          .spyOn(webauthnClient.webauthnState, "matchCredentials")
          .mockReturnValueOnce([{ id: credentialID }]);
      } else {
        jest
          .spyOn(webauthnClient.webauthnState, "matchCredentials")
          .mockReturnValueOnce([]);
      }

      const shouldRegister = await webauthnClient.shouldRegister(user);

      expect(WebauthnSupport.supported).toHaveBeenCalled();
      expect(shouldRegister).toEqual(expected);
    }
  );
});

describe("webauthnClient._createAbortSignal()", () => {
  it("should call abort() on the current controller and return a new one", async () => {
    const signal1 = webauthnClient._createAbortSignal();
    const abortFn = jest.fn();
    webauthnClient.controller.abort = abortFn;
    const signal2 = webauthnClient._createAbortSignal();
    expect(abortFn).toHaveBeenCalled();
    expect(signal1).not.toBe(signal2);
  });
});

describe("webauthnClient.listCredentials()", () => {
  it("should list webauthn credentials", async () => {
    const response = new Response(new XMLHttpRequest());
    response.ok = true;
    response._decodedJSON = [
      {
        id: credentialID,
        public_key: "",
        attestation_type: "",
        aaguid: "",
        created_at: "",
        transports: [],
      },
    ];

    jest.spyOn(webauthnClient.client, "get").mockResolvedValue(response);
    const list = await webauthnClient.listCredentials();
    expect(webauthnClient.client.get).toHaveBeenCalledWith(
      "/webauthn/credentials"
    );
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

      jest.spyOn(webauthnClient.client, "get").mockResolvedValueOnce(response);

      const email = webauthnClient.listCredentials();
      await expect(email).rejects.toThrow(error);
    }
  );

  it("should throw error on API communication failure", async () => {
    webauthnClient.client.get = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const user = webauthnClient.listCredentials();
    await expect(user).rejects.toThrowError("Test error");
  });
});

describe("webauthnClient.updateCredential()", () => {
  it("should update a webauthn credential", async () => {
    const response = new Response(new XMLHttpRequest());
    response.ok = true;

    jest.spyOn(webauthnClient.client, "patch").mockResolvedValue(response);
    const update = await webauthnClient.updateCredential(
      credentialID,
      "new name"
    );
    expect(webauthnClient.client.patch).toHaveBeenCalledWith(
      `/webauthn/credentials/${credentialID}`,
      { name: "new name" }
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

      jest
        .spyOn(webauthnClient.client, "patch")
        .mockResolvedValueOnce(response);

      const email = webauthnClient.updateCredential(credentialID, "new name");
      await expect(email).rejects.toThrow(error);
    }
  );

  it("should throw error on API communication failure", async () => {
    webauthnClient.client.patch = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const user = webauthnClient.updateCredential(credentialID, "new name");
    await expect(user).rejects.toThrowError("Test error");
  });
});

describe("webauthnClient.delete()", () => {
  it("should delete a webauthn credential", async () => {
    const response = new Response(new XMLHttpRequest());
    response.ok = true;

    jest.spyOn(webauthnClient.client, "delete").mockResolvedValue(response);
    const deleteResponse = await webauthnClient.deleteCredential(credentialID);
    expect(webauthnClient.client.delete).toHaveBeenCalledWith(
      `/webauthn/credentials/${credentialID}`
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

      jest
        .spyOn(webauthnClient.client, "delete")
        .mockResolvedValueOnce(response);

      const deleteResponse = webauthnClient.deleteCredential(credentialID);
      await expect(deleteResponse).rejects.toThrow(error);
    }
  );

  it("should throw error on API communication failure", async () => {
    webauthnClient.client.delete = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const user = webauthnClient.deleteCredential(credentialID);
    await expect(user).rejects.toThrowError("Test error");
  });
});
