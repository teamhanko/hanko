import type { Locator, Page } from "@playwright/test";
import { MailSlurper } from "../helper/MailSlurper.js";
import { BasePage } from "./BasePage.js";
import Endpoints from "../helper/Endpoints.js";

export class LoginPasscode extends BasePage {
  readonly signInButton: Locator;
  readonly sendNewCodeLink: Locator;
  readonly headline: Locator;
  readonly mailSlurperClient: MailSlurper;
  readonly passcodeInputs: Locator;

  constructor(page: Page, mailSlurperClient: MailSlurper) {
    super(page);
    this.passcodeInputs = page.locator("input[name*='passcode']");
    this.mailSlurperClient = mailSlurperClient;
    this.signInButton = page.locator("button[type=submit]", {
      hasText: "Sign in",
    });
    this.sendNewCodeLink = page.locator("button", {
      hasText: "Send new code",
    });
    this.headline = page.locator("h1", { hasText: "Enter passcode" });
  }

  async signInWithPasscodeFor(email: string, passcode?: string) {
    // Waiting is discouraged, but we need to give Mailslurper some time to
    // process inbound messages.
    await this.page.waitForTimeout(1000);

    if (!passcode) {
      passcode = await this.mailSlurperClient.getPasscodeFromMostRecentMail(
        email
      );
    }
    const digits = passcode.split("");

    await Promise.all([
      this.page.waitForResponse(Endpoints.API.PASSCODE_LOGIN_FINALIZE),
      this.submitPasscode(digits),
    ]);
  }

  async submitPasscode(digits: string[]) {
    for (let i = 0; i < 6; ++i)
      await this.page.locator(`input[name=passcode${i}]`).fill(digits[i]);
  }

  async sendNewCode() {
    // Introduce some artificial timeout before sending so that Mailslurper does
    // apply the same 'dateSent' values to inbound mails and ordering does not
    // get messed up when retrieving mails via the API.
    await this.page.waitForTimeout(1000);

    const [response] = await Promise.all([
      this.page.waitForResponse(Endpoints.API.PASSCODE_LOGIN_INITIALIZE),
      this.sendNewCodeLink.click(),
    ]);

    return await response.json();
  }
}
