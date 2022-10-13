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

    jest
      .spyOn(webauthnClient.client, "post")
      .mockResolvedValueOnce(initResponse)
      .mockResolvedValueOnce(finalResponse);

    jest.spyOn(webauthnClient.state, "read");
    jest.spyOn(webauthnClient.state, "addCredential");
    jest.spyOn(webauthnClient.state, "write");

    await webauthnClient.login(userID, true);

    expect(webauthnClient._getCredential).toHaveBeenCalledWith({
      ...fakeRequestOptions,
      mediation: "conditional",
    });
    expect(webauthnClient.state.read).toHaveBeenCalledTimes(1);
    expect(webauthnClient.state.addCredential).toHaveBeenCalledWith(
      userID,
      credentialID
    );
    expect(webauthnClient.state.write).toHaveBeenCalledTimes(1);
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

    jest
      .spyOn(webauthnClient.client, "post")
      .mockResolvedValueOnce(initResponse)
      .mockResolvedValueOnce(finalResponse);

    jest.spyOn(webauthnClient.state, "read");
    jest.spyOn(webauthnClient.state, "addCredential");
    jest.spyOn(webauthnClient.state, "write");

    await webauthnClient.register();

    expect(webauthnClient._createCredential).toHaveBeenCalledWith({
      ...fakeCreationOptions,
    });
    expect(webauthnClient.state.read).toHaveBeenCalledTimes(1);
    expect(webauthnClient.state.addCredential).toHaveBeenCalledWith(
      userID,
      credentialID
    );
    expect(webauthnClient.state.write).toHaveBeenCalledTimes(1);
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
    isPlatformAuthenticatorAvailable | userHasCredential | credentialMatched | expected
    ${false}                         | ${false}          | ${false}          | ${false}
    ${true}                          | ${false}          | ${false}          | ${true}
    ${true}                          | ${true}           | ${false}          | ${true}
    ${true}                          | ${true}           | ${true}           | ${false}
  `(
    "should determine correctly if a WebAuthn credential should be registered",
    async ({
      isPlatformAuthenticatorAvailable,
      userHasCredential,
      credentialMatched,
      expected,
    }) => {
      jest
        .spyOn(WebauthnSupport, "isPlatformAuthenticatorAvailable")
        .mockResolvedValueOnce(isPlatformAuthenticatorAvailable);

      const user: User = {
        id: userID,
        email: userID,
        webauthn_credentials: [],
      };

      if (userHasCredential) {
        user.webauthn_credentials.push({ id: credentialID });
      }

      if (credentialMatched) {
        jest
          .spyOn(webauthnClient.state, "matchCredentials")
          .mockReturnValueOnce([{ id: credentialID }]);
      } else {
        jest
          .spyOn(webauthnClient.state, "matchCredentials")
          .mockReturnValueOnce([]);
      }

      const shouldRegister = await webauthnClient.shouldRegister(user);

      expect(
        WebauthnSupport.isPlatformAuthenticatorAvailable
      ).toHaveBeenCalled();
      expect(shouldRegister).toEqual(expected);
    }
  );

  describe("webauthnClient._abortPendingGetCredentialRequest()", () => {
    it("should abort the promise", async () => {
      const controller = new AbortController();
      webauthnClient._getCredentialController = controller;

      jest.spyOn(controller, "abort");

      expect(await webauthnClient._abortPendingGetCredentialRequest()).toEqual(
        undefined
      );
      expect(controller.abort).toHaveBeenCalled();
    });
  });
});
