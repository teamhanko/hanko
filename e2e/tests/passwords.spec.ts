import { test, expect } from "../fixtures/Pages.js";
import { faker } from "@faker-js/faker";
import Endpoints from "../helper/Endpoints.js";

test.describe("@pw", () => {
  test.beforeEach(async ({}) => {
    faker.seed();
  });

  test.describe("@webauthn", () => {
    test.use({
      webauthn: {
        enabled: true,
      },
    });

    test("Register, set up password, skip Passkey, logout, login with password, register Passkey", async ({
      loginEmailPage,
      registerConfirmPage,
      loginPasscodePage,
      loginPasswordPage,
      registerPasswordPage,
      registerAuthenticatorPage,
      securedContentPage,
    }) => {
      const email = faker.internet.email();
      const password = faker.internet.password();

      await test.step("When I visit the baseURL, the LoginEmail page should be shown", async () => {
        await expect(loginEmailPage.headline).toBeVisible();
        await expect(loginEmailPage.signInPasskeyButton).toBeVisible();
      });

      await test.step("And when I submit an email address", async () => {
        await loginEmailPage.continueUsingEmail(email);
      });

      await test.step("The RegisterConfirm page should be shown", async () => {
        await expect(registerConfirmPage.headline).toBeVisible();
      });

      await test.step("And when I confirm the registration", async () => {
        await registerConfirmPage.confirmRegistration();
      });

      await test.step("The LoginPasscode page should be shown", async () => {
        await expect(loginPasscodePage.headline).toBeVisible();
      });

      await test.step("And I should receive a passcode email", async () => {
        const mails = await loginPasscodePage.mailSlurperClient.getMails(email);
        await expect(mails.mailItems).toHaveLength(1);
      });

      await test.step("And when I submit the passcode", async () => {
        await loginPasscodePage.signInWithPasscodeFor(email);
      });

      await test.step("The RegisterPassword page should be shown", async () => {
        await expect(registerPasswordPage.headline).toBeVisible();
      });

      await test.step("And a cookie should be set", async () => {
        await expect(registerPasswordPage).toHaveCookie();
      });

      await test.step("And when I set up a password", async () => {
        await registerPasswordPage.submitPassword(password);
      });

      await test.step("The RegisterAuthenticator page should be shown", async () => {
        await expect(registerAuthenticatorPage.headline).toBeVisible();
      });

      await test.step("And when I skip WebAuthn credential registration", async () => {
        await registerAuthenticatorPage.continue();
      });

      await test.step("The SecuredContent page should be shown", async () => {
        await registerAuthenticatorPage.page.waitForURL(
          Endpoints.APP.SECURED_CONTENT
        );

        await expect(securedContentPage.logoutLink).toBeVisible();
      });

      await test.step("And when I log out", async () => {
        await securedContentPage.logout();
      });

      await test.step("The LoginEmail page should be shown", async () => {
        await expect(loginEmailPage.headline).toBeVisible();
      });

      await test.step("And when I log in with the previously set password", async () => {
        await loginEmailPage.continueUsingEmail(email);
        await expect(loginPasswordPage.headline).toBeVisible();
        await loginPasswordPage.submitPassword(password);
      });

      await test.step("The RegisterAuthenticator page should be shown again", async () => {
        await expect(registerAuthenticatorPage.headline).toBeVisible();
      });

      await test.step("And a cookie should be set", async () => {
        await expect(registerAuthenticatorPage).toHaveCookie();
      });

      await test.step("And when I register a WebAuthn credential", async () => {
        await registerAuthenticatorPage.registerPasskey();
      });

      await test.step("The SecuredContent page should be shown", async () => {
        await registerAuthenticatorPage.page.waitForURL(
          Endpoints.APP.SECURED_CONTENT
        );

        await expect(securedContentPage.logoutLink).toBeVisible();
      });
    });
  });

  test.describe("@nowebauthn", () => {
    test("Password recovery", async ({
      loginEmailPage,
      registerConfirmPage,
      loginPasscodePage,
      loginPasswordPage,
      registerPasswordPage,
      securedContentPage,
    }) => {
      const email = faker.internet.email();
      const password = faker.internet.password();

      await test.step("When I visit the baseURL, the LoginEmail page should be shown", async () => {
        await expect(loginEmailPage.headline).toBeVisible();
        await expect(loginEmailPage.signInPasskeyButton).toBeHidden();
      });

      await test.step("And when I submit an email address", async () => {
        await loginEmailPage.continueUsingEmail(email);
      });

      await test.step("The RegisterConfirm page should be shown", async () => {
        await expect(registerConfirmPage.headline).toBeVisible();
      });

      await test.step("And when I confirm the registration", async () => {
        await registerConfirmPage.confirmRegistration();
      });

      await test.step("The LoginPasscode page should be shown", async () => {
        await expect(loginPasscodePage.headline).toBeVisible();
      });

      await test.step("And I should receive a passcode email", async () => {
        const mails = await loginPasscodePage.mailSlurperClient.getMails(email);
        await expect(mails.mailItems).toHaveLength(1);
      });

      await test.step("And when I submit the passcode", async () => {
        await loginPasscodePage.signInWithPasscodeFor(email);
      });

      await test.step("The RegisterPassword page should be shown", async () => {
        await expect(registerPasswordPage.headline).toBeVisible();
      });

      await test.step("And a cookie should be set", async () => {
        await expect(registerPasswordPage).toHaveCookie();
      });

      await test.step("And when I set up a password", async () => {
        await registerPasswordPage.submitPassword(password);
      });

      await test.step("The SecuredContent page should be shown", async () => {
        await registerPasswordPage.page.waitForURL(
          Endpoints.APP.SECURED_CONTENT
        );

        await expect(securedContentPage.logoutLink).toBeVisible();
      });

      await test.step("And when I log out", async () => {
        await securedContentPage.logout();
      });

      await test.step("The LoginEmail page should be shown", async () => {
        await expect(loginEmailPage.headline).toBeVisible();
      });

      await test.step("And when I submit my email address again", async () => {
        await loginEmailPage.continueUsingEmail(email);
      });

      await test.step("The LoginPassword page should be shown", async () => {
        await expect(loginPasswordPage.headline).toBeVisible();
        await expect(loginPasswordPage.forgotPasswordLink).toBeVisible();
      });

      await test.step("And when I trigger password recovery", async () => {
        await loginPasswordPage.recovery();
      });

      await test.step("The LoginPasscode page should be shown", async () => {
        await expect(loginPasscodePage.headline).toBeVisible();
      });

      await test.step("And I should receive another passcode email", async () => {
        const mails = await loginPasscodePage.mailSlurperClient.getMails(email);
        await expect(mails.mailItems).toHaveLength(2);
      });

      await test.step("And when I submit the passcode", async () => {
        await loginPasscodePage.signInWithPasscodeFor(email);
      });

      await test.step("The RegisterPassword page should be shown", async () => {
        await expect(registerPasswordPage.headline).toBeVisible();
      });

      await test.step("And a cookie should have been set", async () => {
        await expect(registerPasswordPage).toHaveCookie();
      });

      const newPassword = faker.internet.password();
      await test.step("And when I set a new password", async () => {
        await registerPasswordPage.submitPassword(newPassword);
      });

      await test.step("The SecuredContent page should be shown", async () => {
        await expect(securedContentPage.logoutLink).toBeVisible();
      });

      await test.step("And when I log out and log in with the old password", async () => {
        await securedContentPage.logout();
        await loginEmailPage.continueUsingEmail(email);
        await loginPasswordPage.submitPassword(password);
      });

      await test.step("An invalid credentials error should be shown", async () => {
        await expect(loginPasswordPage.errorMessage).toHaveText(
          "Wrong email or password."
        );
      });

      await test.step("And when I log in with the new password", async () => {
        await loginPasswordPage.submitPassword(newPassword);
      });

      await test.step("The SecuredContent page should be shown", async () => {
        await expect(securedContentPage.logoutLink).toBeVisible();
      });

      await test.step("And a cookie should have been set", async () => {
        await expect(securedContentPage).toHaveCookie();
      });
    });
  });
});
