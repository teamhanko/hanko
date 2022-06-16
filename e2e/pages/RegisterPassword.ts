import type { Locator, Page } from "@playwright/test";
import { BasePage } from "./BasePage.js";
import Endpoints from "../helper/Endpoints.js";

export class RegisterPassword extends BasePage {
  readonly passwordInput: Locator;
  readonly continueButton: Locator;
  readonly headline: Locator;

  constructor(page: Page) {
    super(page);
    this.passwordInput = page.locator("input[name=password]");
    this.continueButton = page.locator("button[type=submit]", {
      hasText: "Continue",
    });
    this.headline = page.locator("h1", { hasText: "Set new password" });
  }

  async submitPassword(password: string) {
    await this.passwordInput.fill(password);
    await Promise.all([
      this.page.waitForResponse(Endpoints.API.PASSWORD),
      this.continueButton.click(),
    ]);
  }
}
