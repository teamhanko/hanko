import type { Locator, Page } from "@playwright/test";
import { BasePage } from "./BasePage.js";
import Endpoints from "../helper/Endpoints.js";

export class LoginEmail extends BasePage {
  readonly emailInput: Locator;
  readonly continueButton: Locator;
  readonly signInPasskeyButton: Locator;
  readonly headline: Locator;

  constructor(page: Page) {
    super(page);
    this.emailInput = page.locator("input[name=email]");
    this.continueButton = page.locator("button[type=submit]", {
      hasText: "Continue",
    });
    this.signInPasskeyButton = page.locator("button[type=submit]", {
      hasText: "Sign in with passkey",
    });
    this.headline = page.locator("h1", { hasText: "Sign in or sign up" });
  }

  async continueUsingEmail(email: string) {
    await this.emailInput.fill(email);
    await Promise.all([
      this.page.waitForResponse(Endpoints.API.USER),
      this.continueButton.click(),
    ]);
  }

  async signInWithPasskey() {
    await Promise.all([
      this.page.waitForResponse(Endpoints.API.WEBAUTHN_LOGIN_INITIALIZE),
      this.page.waitForResponse(Endpoints.API.WEBAUTHN_LOGIN_FINALIZE),
      this.signInPasskeyButton.click(),
    ]);
  }
}
