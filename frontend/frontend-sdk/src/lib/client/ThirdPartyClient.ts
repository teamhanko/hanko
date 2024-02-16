import { Client } from "./Client";
import { ThirdPartyError } from "../Errors";

/**
 * A class that handles communication with the Hanko API for the purposes
 * of authenticating through a third party provider.
 *
 * @constructor
 * @category SDK
 * @subcategory Clients
 * @extends {Client}
 */
export class ThirdPartyClient extends Client {
  /**
   * Performs a request to the Hanko API that redirects to the given
   * third party provider.
   *
   * @param {string} provider - The name of the third party provider
   * @param {string} redirectTo - The URL to redirect to after a successful third party authentication
   * @throws {ThirdPartyError}
   * @see http://docs.hanko.io/api/public#tag/Third-Party/operation/thirdPartyAuth
   */
  async auth(provider: string, redirectTo: string): Promise<void> {
    const url = new URL("/thirdparty/auth", this.client.api);

    if (!provider) {
      throw new ThirdPartyError(
        "somethingWentWrong",
        new Error("provider missing from request")
      );
    }

    if (!redirectTo) {
      throw new ThirdPartyError(
        "somethingWentWrong",
        new Error("redirectTo missing from request")
      );
    }

    url.searchParams.append("provider", provider);
    url.searchParams.append("redirect_to", redirectTo);

    window.location.assign(url.href);
  }

  /**
   * Get a third party error from the current location's query params.
   * @returns {(ThirdPartyError|undefined)} The ThirdPartyError.
   */
  getError() {
    const params = new URLSearchParams(window.location.search);
    const error = params.get("error");
    const errorDescription = params.get("error_description");
    if (error) {
      let code = "";
      switch (error) {
        case "access_denied":
          code = "thirdPartyAccessDenied";
          break;
        case "user_conflict":
          code = "emailAddressAlreadyExistsError";
          break;
        case "multiple_accounts":
          code = "thirdPartyMultipleAccounts";
          break;
        case "unverified_email":
          code = "thirdPartyUnverifiedEmail";
          break;
        case "email_maxnum":
          code = "maxNumOfEmailAddressesReached";
          break;
        case "signup_disabled":
          code = "signupDisabled";
          break;
        default:
          code = "somethingWentWrong";
      }

      return new ThirdPartyError(code, new Error(errorDescription));
    }
  }
}
