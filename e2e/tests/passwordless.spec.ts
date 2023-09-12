import { test, expect } from "../fixtures/Pages.js";
import { faker } from "@faker-js/faker";
import Endpoints from "../helper/Endpoints.js";
import Accounts from "../helper/Accounts.js";

test.describe("@nopw", () => {
  test.beforeEach(async ({}) => {
    faker.seed();
  });

  test.describe("@webauthn", () => {
    const transports = ["internal", "usb", "nfc", "ble"];
    for (const transport of transports) {
      test.use({
        webauthn: {
          enabled: true,
          authenticator: {
            transport: transport,
          },
        },
      });

      test(`Register, add passkey, logout, login with passkey with authenticator transport ${transport}`, async ({
        loginEmailPage,
        registerConfirmPage,
        loginPasscodePage,
        registerAuthenticatorPage,
        securedContentPage,
      }) => {
        const email = faker.internet.email();

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
          const mails = await loginPasscodePage.mailSlurperClient.getMails(
            email
          );
          await expect(mails.mailItems).toHaveLength(1);
        });

        await test.step("And when I submit the passcode", async () => {
          await loginPasscodePage.signInWithPasscodeFor(email);
        });

        await test.step("The RegisterAuthenticator page should be shown", async () => {
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

        await test.step("And when I log out", async () => {
          await securedContentPage.logout();
        });

        await test.step("The LoginEmail page should be shown", async () => {
          await expect(loginEmailPage.headline).toBeVisible();
        });

        await test.step("And when I login with the previously registered WebAuthn credential", async () => {
          await loginEmailPage.signInWithPasskey();
        });

        await test.step("The SecuredContent page should be shown", async () => {
          await expect(securedContentPage.logoutLink).toBeVisible();
        });
      });
    }
  });

  test.describe("@nowebauthn", () => {
    test("Register, login with passcode", async ({
      loginEmailPage,
      loginPasscodePage,
      registerConfirmPage,
      registerAuthenticatorPage,
      securedContentPage,
    }) => {
      const email = faker.internet.email();

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

      await test.step("The RegisterAuthenticator page should be shown", async () => {
        await expect(registerAuthenticatorPage.headline).toBeVisible();
      });

      await test.step("And a cookie should be set", async () => {
        await expect(registerAuthenticatorPage).toHaveCookie();
      });

      await test.step("And when I skip WebAuthn credential registration", async () => {
        await registerAuthenticatorPage.skip();
      });

      await test.step("The SecuredContent page should be shown", async () => {
        await expect(securedContentPage.logoutLink).toBeVisible();
      });

      await test.step("And a cookie should have been set", async () => {
        await expect(securedContentPage).toHaveCookie();
      });
    });

    test("Logging in with existing user will prompt for passcode", async ({
      loginEmailPage,
      loginPasscodePage
    }) => {
      const email = Accounts.test.email;

      await test.step("When I visit the baseURL, the LoginEmail page should be shown", async () => {
        await expect(loginEmailPage.headline).toBeVisible();
      });

      await test.step("And when I submit an email address", async () => {
        await loginEmailPage.continueUsingEmail(email);
      });

      await test.step("The LoginPasscode page should be shown", async () => {
        await expect(loginPasscodePage.headline).toBeVisible();
      });
    });
  });
});
