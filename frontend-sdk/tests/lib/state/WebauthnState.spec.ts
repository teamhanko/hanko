import { decodedLSContent } from "../../setup";
import { WebauthnState } from "../../../src/lib/state/WebauthnState";
import { Credential } from "../../../src";

describe("webauthnState.read()", () => {
  it("should read the webauthn state", async () => {
    const state = new WebauthnState();

    expect(state.read()).toEqual(state);
  });
});

describe("webauthnState.getCredentials()", () => {
  it("should read the webauthn state", async () => {
    const ls = decodedLSContent();
    const state = new WebauthnState();
    const userID = Object.keys(ls.users)[0];

    state.ls = ls;
    expect(state.getCredentials(userID)).toEqual(
      ls.users[userID].webauthn.credentials
    );
  });
});

describe("webauthnState.addCredential()", () => {
  it("should add a credential id", async () => {
    const ls = decodedLSContent();
    const state = new WebauthnState();
    const userID = Object.keys(ls.users)[0];
    const credentialID = "testCredentialID";

    expect(state.addCredential(userID, credentialID)).toEqual(state);
    expect(state.ls.users[userID].webauthn.credentials).toContainEqual(
      credentialID
    );
  });
});

describe("webauthnState.matchCredentials()", () => {
  it("should match credential ids", async () => {
    const ls = decodedLSContent();
    const state = new WebauthnState();
    const userID = Object.keys(ls.users)[0];
    const credentials = ls.users[userID].webauthn.credentials.map(
      (id) => ({ id } as Credential)
    );
    const more = [{ id: "testCredentialID" } as Credential];

    state.ls = ls;

    expect(state.matchCredentials(userID, credentials.concat(more))).toEqual(
      credentials
    );
  });

  it("shouldn't match credential ids", async () => {
    const ls = decodedLSContent();
    const state = new WebauthnState();
    const userID = Object.keys(ls.users)[0];
    const credentials = ls.users[userID].webauthn.credentials.map(
      (id) => ({ id } as Credential)
    );

    state.ls = ls;
    state.ls.users[userID].webauthn.credentials = ["testCredentialID"];

    expect(state.matchCredentials(userID, credentials)).toEqual([]);
  });
});
