import type { Locator, Page } from "@playwright/test";
import { BasePage } from "./BasePage.js";
import { expect } from "../fixtures/Pages.js";

export class NoAccountFound extends BasePage {
  readonly backLink: Locator;
  readonly signUpButton: Locator;
  readonly headline: Locator;

  constructor(page: Page) {
    super(page);
    this.backLink = page.locator("button", { hasText: "Back" });
    this.signUpButton = page.locator("button[type=submit]", {
      hasText: "Sign up",
    });
    this.headline = page.locator("h1", { hasText: "No account found" });
  }

  async assertSignupButtonNotVisible() {
    await expect(this.signUpButton).not.toBeVisible();
  }

  async assertNoAccountFoundText(email: string) {
    const text = this.page.locator("p", {hasText: `No account exists for "${email}".`});
    await expect(text).toBeVisible();
  }

  async back() {
    await this.backLink.click();
  }
}
