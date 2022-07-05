import { BasePage } from "./BasePage.js";
import type { Locator, Page } from "@playwright/test";
import Endpoints from "../helper/Endpoints.js";

export class Error extends BasePage {
  readonly headline: Locator;
  readonly continueButton: Locator;

  constructor(page: Page) {
    super(page);
    this.headline = page.locator("h1", {
      hasText: "An error has occurred",
    });
    this.continueButton = page.locator("button[type=submit]", {
      hasText: "Continue",
    });
  }

  async continue() {
    await Promise.all([
      this.page.waitForResponse(Endpoints.API.WELL_KNOWN_CONFIG),
      this.continueButton.click(),
    ]);
  }
}
