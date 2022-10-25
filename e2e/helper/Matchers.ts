import { BasePage } from "../pages/BasePage.js";

export async function toHaveCookie(received: BasePage, name = "hanko") {
  if (typeof received.hasCookie !== "function") {
    return {
      message: () => `matcher only applicable to type BasePage`,
      pass: false,
    };
  }

  const pass = await received.hasCookie(name);
  if (pass) {
    return {
      message: () => `cookie with '${name}' present`,
      pass: true,
    };
  } else {
    return {
      message: () => `no cookie with name '${name}' present`,
      pass: false,
    };
  }
}

export async function toHaveLocalStorageEntry(
  received: BasePage,
  origin = "http://localhost:8888",
  name = "hanko"
) {
  if (typeof received.getLocalStorageValue !== "function") {
    return {
      message: () => `matcher only applicable to type BasePage`,
      pass: false,
    };
  }

  const localStorageEntry = await received.getLocalStorageValue(origin, name);

  if (localStorageEntry) {
    return {
      message: () =>
        `local storage entry for origin '${origin}' with key '${name}' present`,
      pass: true,
    };
  } else {
    return {
      message: () =>
        `local storage entry for origin '${origin}' with key '${name}' not present`,
      pass: false,
    };
  }
}

export async function toHaveLocalStorageEntryForUserWithCredential(
  received: BasePage,
  userId: string,
  credentialId: string,
  origin = "http://localhost:8888",
  name = "hanko"
) {
  if (typeof received.getDecodedLocalStorageValue !== "function") {
    return {
      message: () => `matcher only applicable to type BasePage`,
      pass: false,
    };
  }

  const { users } = await received.getDecodedLocalStorageValue(origin, name);

  if (!users) {
    return {
      message: () =>
        `credential '${credentialId}' for user '${userId}' present`,
      pass: false,
    };
  }

  const userState = users[userId];
  const pass = userState.webauthn.credentials.includes(credentialId);

  if (pass) {
    return {
      message: () =>
        `credential '${credentialId}' for user '${userId}' present`,
      pass: true,
    };
  } else {
    return {
      message: () =>
        `credential '${credentialId}' for user '${userId}' not present`,
      pass: false,
    };
  }
}

export async function toHaveLocalStorageEntryForUserWithPasscode(
  received: BasePage,
  userId: string,
  passcodeId: string,
  origin = "http://localhost:8888",
  name = "hanko"
) {
  if (typeof received.getDecodedLocalStorageValue !== "function") {
    return {
      message: () => `matcher only applicable to type BasePage`,
      pass: false,
    };
  }

  const { users } = await received.getDecodedLocalStorageValue(origin, name);

  if (!users) {
    return {
      message: () => `local storage value not present`,
      pass: false,
    };
  }

  const userState = users[userId];

  if (!userState) {
    return {
      message: () => `user state for user ${userId} not present`,
      pass: false,
    };
  }

  const pass = userState.passcode.id === passcodeId;

  if (pass) {
    return {
      message: () => `passcode '${passcodeId}' for user '${userId}' present`,
      pass: true,
    };
  } else {
    return {
      message: () =>
        `passcode '${passcodeId}' for user '${userId}' not present`,
      pass: false,
    };
  }
}
