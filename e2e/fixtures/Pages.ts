import { expect, test as base } from "@playwright/test";
import { LoginEmail } from "../pages/LoginEmail.js";
import { RegisterConfirm } from "../pages/RegisterConfirm.js";
import { LoginPasscode } from "../pages/LoginPasscode.js";
import { RegisterAuthenticator } from "../pages/RegisterAuthenticator.js";
import { SecuredContent } from "../pages/SecuredContent.js";
import { LoginPassword } from "../pages/LoginPassword.js";
import { RegisterPassword } from "../pages/RegisterPassword.js";
import { MailSlurper } from "../helper/MailSlurper.js";
import * as Matchers from "../helper/Matchers.js";
import { Error } from "../pages/Error.js";
import Endpoints from "../helper/Endpoints.js";
import Setup from "../helper/Setup.js";

export type Pages = {
  errorPage: Error;
  loginEmailPage: LoginEmail;
  registerConfirmPage: RegisterConfirm;
  loginPasscodePage: LoginPasscode;
  loginPasswordPage: LoginPassword;
  registerPasswordPage: RegisterPassword;
  registerAuthenticatorPage: RegisterAuthenticator;
  securedContentPage: SecuredContent;
};

export type AuthenticatorOptions = {
  protocol?: string;
  transport?: string;
  hasResidentKey?: boolean;
  hasUserVerification?: boolean;
  isUserVerified?: boolean;
};

export type WebAuthnOptions = {
  enabled: boolean;
  authenticator?: AuthenticatorOptions;
};

export type TestOptions = {
  webauthn: WebAuthnOptions;
};

export const test = base.extend<TestOptions & Pages>({
  webauthn: [{ enabled: false }, { option: true }],

  errorPage: async ({ page }, use) => {
    await use(new Error(page));
  },

  loginEmailPage: async ({ baseURL, page, webauthn }, use) => {
    await Setup.webauthn(page, webauthn);

    await Promise.all([
      page.waitForResponse(Endpoints.API.WELL_KNOWN_CONFIG),
      page.goto(baseURL!),
    ]);

    const loginEmailPage: LoginEmail = new LoginEmail(page);
    await use(loginEmailPage);
  },

  registerConfirmPage: async ({ page }, use) => {
    await use(new RegisterConfirm(page));
  },

  loginPasscodePage: async ({ page }, use) => {
    const mail = new MailSlurper();
    const loginPasscode = new LoginPasscode(page, mail);
    await use(loginPasscode);
  },

  loginPasswordPage: async ({ page }, use) => {
    await use(new LoginPassword(page));
  },

  registerAuthenticatorPage: async ({ page }, use) => {
    await use(new RegisterAuthenticator(page));
  },

  registerPasswordPage: async ({ page }, use) => {
    await use(new RegisterPassword(page));
  },

  securedContentPage: async ({ page }, use) => {
    await use(new SecuredContent(page));
  },
});

expect.extend({
  ...Matchers,
});

export { expect };
