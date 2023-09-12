import { test, expect } from "../fixtures/Pages.js";
import { faker } from "@faker-js/faker";
import Accounts from "../helper/Accounts.js";

test.describe("@nosignup", () => {
    test("Login with account that does not exist", async ({
        loginEmailNoSignupPage,
        noAccountFoundPage
    }) => {
        const email = faker.internet.email();

        await test.step("When I visit the baseURL, the LoginEmailNoSignup page should be shown", async () => {
            await expect(loginEmailNoSignupPage.headline).toBeVisible();
            await expect(loginEmailNoSignupPage.signInPasskeyButton).toBeVisible();
        });

        await test.step("And when I submit an email address", async () => {
            await loginEmailNoSignupPage.continueUsingEmail(email);
        });

        await test.step("No account should be found", async () => {
            await noAccountFoundPage.assertNoAccountFoundText(email);
        });

        await test.step("Signup button should not be visible", async() => {
            await noAccountFoundPage.assertSignupButtonNotVisible();
        });

        await test.step("Navigating back should take me back to LoginEmailNoSignup page", async () => {
            await noAccountFoundPage.back();
            await expect(loginEmailNoSignupPage.headline).toBeVisible();
            await expect(loginEmailNoSignupPage.signInPasskeyButton).toBeVisible();
        });
    });

    test("Login with existing account", async ({
        loginEmailNoSignupPage,
        loginPasscodePage
    }) => {
        const email = Accounts.test.email;

        await test.step("When I visit the baseURL, the LoginEmailNoSignup page should be shown", async () => {
            await expect(loginEmailNoSignupPage.headline).toBeVisible();
            await expect(loginEmailNoSignupPage.signInPasskeyButton).toBeVisible();
        });

        await test.step("And when I submit an email that already exists", async () => {
            await loginEmailNoSignupPage.continueUsingEmail(email);
        });

        await test.step("Login passocde page should be displayed", async () => {
            await expect(loginPasscodePage.headline).toBeVisible();
        });
    });
});
