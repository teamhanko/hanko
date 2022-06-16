import { test, expect } from "../fixtures/Pages.js";
import { LoginEmail } from "../pages/LoginEmail.js";
import type { Mails } from "../helper/MailSlurper.js";
import Endpoints from "../helper/Endpoints.js";
import { faker } from "@faker-js/faker";

test.describe("@common", () => {
  test.describe("@nowebauthn", () => {
    test("Error page on faulty config", async ({
      baseURL,
      page,
      errorPage,
    }) => {
      await test.step("Given the API returns an error response once when the public config is requested", async () => {
        await page.route(
          Endpoints.API.WELL_KNOWN_CONFIG,
          async (route) => {
            const response = await page.request.fetch(route.request());
            await route.fulfill({
              response,
              status: 500,
              body: "Internal Server error",
            });
          },
          { times: 1 }
        );
      });

      await test.step("When I visit the base URL", async () => {
        await page.goto(baseURL!);
      });

      await test.step("The error page should be shown", async () => {
        await expect(errorPage.headline).toBeVisible();
        await expect(errorPage.errorMessage).toBeVisible();
      });

      await test.step("And when clicking the continue button", async () => {
        await errorPage.continue();
      });

      await test.step("The LoginEmail page should be shown", async () => {
        const loginEmailPage = new LoginEmail(page);
        await expect(loginEmailPage.headline).toBeVisible();
      });
    });

    test("Expired passcode", async ({
      page,
      loginEmailPage,
      registerConfirmPage,
      loginPasscodePage,
    }) => {
      const email = faker.internet.email();

      await test.step("Given the API creates a passcode that expires immediately", async () => {
        await page.route(
          Endpoints.API.PASSCODE_LOGIN_INITIALIZE,
          async (route) => {
            const response = await page.request.fetch(route.request());
            const body = await response.json();
            await route.fulfill({
              response,
              body: JSON.stringify({ ...body, ttl: 0 }),
            });
          },
          { times: 1 }
        );
      });

      await test.step("When I submit an email address", async () => {
        await loginEmailPage.continueUsingEmail(email);
      });

      await test.step("The RegisterConfirm page should be shown", async () => {
        await expect(registerConfirmPage.headline).toBeVisible();
      });

      await test.step("And when I confirm the registration", async () => {
        await registerConfirmPage.confirmRegistration();
      });

      await test.step("The passcode login page should be shown", async () => {
        await expect(loginPasscodePage.headline).toBeVisible();
      });

      await test.step("And a passcode expiry message should be shown", async () => {
        await expect(loginPasscodePage.errorMessage).toHaveText(/expired/);
      });

      await test.step("And input elements should be disabled", async () => {
        await expect(loginPasscodePage.signInButton).toBeDisabled();
        for (let i = 0; i < 6; ++i)
          await expect(
            loginPasscodePage.page.locator(`input[name=passcode${i}]`)
          ).toBeDisabled();
      });

      await test.step("And when I request a new passcode", async () => {
        await loginPasscodePage.sendNewCode();
      });

      await test.step("The error message should be hidden", async () => {
        await expect(loginPasscodePage.errorMessage).toBeHidden();
      });

      await test.step("And input elements should be enabled", async () => {
        await expect(loginPasscodePage.signInButton).toBeEnabled();
        for (let i = 0; i < 6; ++i)
          await expect(
            loginPasscodePage.page.locator(`input[name=passcode${i}]`)
          ).toBeEnabled();
      });
    });

    test("Requesting a new passcode but logging in with an old passcode fails", async ({
      loginEmailPage,
      loginPasscodePage,
      registerConfirmPage,
    }) => {
      const email = faker.internet.email();

      await test.step("When I visit the baseURL, the LoginEmail page should be shown", async () => {
        await expect(loginEmailPage.headline).toBeVisible();
      });

      await test.step("And when I submit an email address", async () => {
        await loginEmailPage.continueUsingEmail(email);
      });

      await test.step("The RegisterConfirm page should be shown", async () => {
        await expect(registerConfirmPage.headline).toBeVisible();
      });

      let userId: string;
      let passcodeId: string;

      await test.step("And when I confirm the registration", async () => {
        [{ id: userId }, { id: passcodeId }] =
          await registerConfirmPage.confirmRegistration();
      });

      await test.step("The LoginPasscode page should be shown", async () => {
        await expect(loginPasscodePage.headline).toBeVisible();
      });

      await test.step("And I should receive a passcode email", async () => {
        const mails = await loginPasscodePage.mailSlurperClient.getMails(email);
        await expect(mails.mailItems).toHaveLength(1);
      });

      let newPasscodeId: string;

      await test.step("And when I request a new passcode", async () => {
        ({ id: newPasscodeId } = await loginPasscodePage.sendNewCode());
      });

      await test.step("The API responds with a new passcode", async () => {
        await expect(newPasscodeId).not.toBe(passcodeId);
      });

      await test.step("And the old passcode in the local storage user state is replaced", async () => {
        await expect(
          loginPasscodePage
        ).toHaveLocalStorageEntryForUserWithPasscode(userId, newPasscodeId);
      });

      let mails: Mails;

      await test.step("And I receive a new passcode email", async () => {
        mails = await loginPasscodePage.mailSlurperClient.getMails(email);
        await expect(mails.mailItems).toHaveLength(2);
      });

      await test.step("And when I submit the old passcode", async () => {
        const oldPasscode =
          await loginPasscodePage.mailSlurperClient.getPasscodeFromMail(
            mails.mailItems[1]
          );

        await loginPasscodePage.signInWithPasscodeFor(email, oldPasscode);
      });

      await test.step("An error should be shown", async () => {
        await expect(loginPasscodePage.errorMessage).toBeVisible();
      });

      await test.step("And no cookie should be created", async () => {
        await expect(loginPasscodePage).not.toHaveCookie();
      });
    });
  });
});
