import type { Page } from "@playwright/test";
import type { WebAuthnOptions } from "../fixtures/Pages.js";

async function setUpWebAuthn(page: Page, options: WebAuthnOptions) {
  const defaultAuthenticatorOptions = {
    protocol: "ctap2",
    transport: "internal",
    hasResidentKey: true,
    hasUserVerification: true,
    isUserVerified: true,
  };

  if (options.enabled) {
    const client = await page.context().newCDPSession(page);
    await client.send("WebAuthn.enable");
    await client.send("WebAuthn.addVirtualAuthenticator", {
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore
      options: {
        ...defaultAuthenticatorOptions,
        ...options.authenticator,
      },
    });
  }
}

const Setup = {
  webauthn: setUpWebAuthn,
};

export default Setup;
