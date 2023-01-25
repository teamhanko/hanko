import type { Locator, Page } from "@playwright/test";
import { BasePage } from "./BasePage.js";
import Endpoints from "../helper/Endpoints.js";

export class LoginPassword extends BasePage {
  readonly passwordInput: Locator;
  readonly signInButton: Locator;
  readonly backLink: Locator;
  readonly forgotPasswordLink: Locator;
  readonly headline: Locator;

  constructor(page: Page) {
    super(page);
    this.passwordInput = page.locator("input[name=password]");
    this.signInButton = page.locator("button[type=submit]", {
      hasText: "Sign in",
    });
    this.backLink = page.locator("button", { hasText: "Back" });
    this.forgotPasswordLink = page.locator("button", {
      hasText: "Forgot your password?",
    });
    this.headline = page.locator("h1", { hasText: "Enter password" });
  }

  async submitPassword(password: string) {
    await this.passwordInput.fill(password);
    await Promise.all([
      this.page.waitForResponse(Endpoints.API.PASSWORD_LOGIN),
      this.signInButton.click(),
    ]);
  }

  async back() {
    await this.backLink.click();
  }

  async recovery() {
    await Promise.all([
      this.page.waitForResponse(Endpoints.API.PASSCODE_LOGIN_INITIALIZE),
      this.forgotPasswordLink.click(),
    ]);
  }
}
