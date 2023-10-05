import type { Locator, Page } from "@playwright/test";
import { BasePage } from "./BasePage.js";
import Endpoints from "../helper/Endpoints.js";
import { expect } from "../fixtures/Pages.js";

export class RegisterAuthenticator extends BasePage {
  readonly setUpPasskeyButton: Locator;
  readonly skipLink: Locator;
  readonly headline: Locator;

  constructor(page: Page) {
    super(page);
    this.setUpPasskeyButton = page.locator("button[type=submit]", {
      hasText: "Create a passkey",
    });
    this.skipLink = page.locator("button", {
      hasText: "Skip",
    });
    this.headline = page.locator("h1", { hasText: "Create a passkey" });
  }

  async registerPasskey() {
    const [initResponse, finalResponse] = await Promise.all([
      this.page.waitForResponse(Endpoints.API.WEBAUTHN_REGISTRATION_INITIALIZE),
      this.page.waitForResponse(Endpoints.API.WEBAUTHN_REGISTRATION_FINALIZE),
      this.setUpPasskeyButton.click(),
    ]);

    const initResponseJson = await initResponse.json();
    const finalResponseJson = await finalResponse.json();
    const { user_id: userId, credential_id: credentialId } = finalResponseJson;

    await expect(
      this,
      "The credential is encoded in the local storage user state"
    ).toHaveLocalStorageEntryForUserWithCredential(userId, credentialId);

    return [initResponseJson, finalResponseJson];
  }

  async skip() {
    await this.skipLink.click();
  }
}
