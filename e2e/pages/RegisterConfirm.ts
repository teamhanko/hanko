import type { Locator, Page } from "@playwright/test";
import { BasePage } from "./BasePage.js";
import Endpoints from "../helper/Endpoints.js";
import { expect } from "../fixtures/Pages.js";

export class RegisterConfirm extends BasePage {
  readonly backLink: Locator;
  readonly signUpButton: Locator;
  readonly headline: Locator;

  constructor(page: Page) {
    super(page);
    this.backLink = page.locator("a", { hasText: "Back" });
    this.signUpButton = page.locator("button[type=submit]", {
      hasText: "Sign up",
    });
    this.headline = page.locator("h1", { hasText: "Create account?" });
  }

  async confirmRegistration() {
    const [usersResponse, passcodeInitResponse] = await Promise.all([
      this.page.waitForResponse(Endpoints.API.USERS),
      this.page.waitForResponse(Endpoints.API.PASSCODE_LOGIN_INITIALIZE),
      this.signUpButton.click(),
    ]);

    const usersResponseJson = await usersResponse.json();
    const passcodeInitResponseJson = await passcodeInitResponse.json();

    await expect(
      this,
      "The passcode is encoded in the local storage user state"
    ).toHaveLocalStorageEntryForUserWithPasscode(
      usersResponseJson.id,
      passcodeInitResponseJson.id
    );

    return [usersResponseJson, passcodeInitResponseJson];
  }

  async back() {
    await this.backLink.click();
  }
}
