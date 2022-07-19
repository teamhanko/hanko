import type { Locator, Page } from "@playwright/test";
import { BasePage } from "./BasePage.js";
import Endpoints from "../helper/Endpoints.js";
import { expect } from "../fixtures/Pages.js";

export class SecuredContent extends BasePage {
  readonly logoutLink: Locator;

  constructor(page: Page) {
    super(page);
    this.logoutLink = page.locator("a", { hasText: "Logout" });
  }

  async logout() {
    await Promise.all([
      this.page.waitForResponse(Endpoints.APP.LOGOUT),
      this.logoutLink.click(),
    ]);

    await expect(
      this,
      "Logging out should clear the cookie"
    ).not.toHaveCookie();
  }
}
