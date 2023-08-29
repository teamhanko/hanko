import { Client } from "./Client";
import { NotFoundError, TechnicalError, ThirdPartyError } from "../Errors";

/**
 * A class that handles communication with the Hanko API for the purposes
 * of authenticating through a third party provider.
 *
 * @constructor
 * @category SDK
 * @subcategory Clients
 * @extends {Client}
 */
export class EnterpriseClient extends Client {
  /**
   * Extracts the domain from an email address
   * @param {string} email E-Mail address of the user from which the domain will be extracted.
   * @throws {ThirdPartyError}
   * @private
   */
  private getDomain(email: string): string {
    if (!email) {
      throw new ThirdPartyError(
        "somethingWentWrong",
        new Error("email missing from request"),
      );
    }

    const emailParts = email.split("@");
    if (emailParts.length !== 2) {
      throw new ThirdPartyError(
        "somethingWentWrong",
        new Error("email is not in a valid email format."),
      );
    }

    const domain = emailParts[1].trim();
    if (domain === "") {
      throw new ThirdPartyError(
        "somethingWentWrong",
        new Error("email is not in a valid email format."),
      );
    }

    return domain;
  }

  /**
   * Performs a request to the Hanko API to check if there is a provider for the users e-mail domain
   *
   * @param {string} email - E-Mail address of the user to login
   */
  async hasProvider(email: string): Promise<boolean> {
    const domain = this.getDomain(email);

    return this.client.get(`/saml/provider?domain=${domain}`).then((resp) => {
      if (resp.status == 404) {
        throw new NotFoundError(new Error("provider not found"));
      }

      if (!resp.ok) {
        throw new TechnicalError(new Error("unable to fetch provider"));
      }

      return resp.ok;
    });
  }

  /**
   * Performs a request to the Hanko API that redirects to the given
   * third party provider.
   *
   * @param {string} email - E-Mail address of the user
   * @param {string} redirectTo - The URL to redirect to after a successful third party authentication
   * @throws {ThirdPartyError}
   * @see http://docs.hanko.io/api/public#tag/Third-Party/operation/enterpriseAuth
   */
  auth(email: string, redirectTo: string): void {
    const url = new URL("/saml/auth", this.client.api);
    const domain = this.getDomain(email);

    if (!redirectTo) {
      throw new ThirdPartyError(
        "somethingWentWrong",
        new Error("redirectTo missing from request"),
      );
    }

    url.searchParams.append("domain", domain);
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
      let code;
      switch (error) {
        case "access_denied":
          code = "enterpriseAccessDenied";
          break;
        case "user_conflict":
          code = "emailAddressAlreadyExistsError";
          break;
        case "multiple_accounts":
          code = "enterpriseMultipleAccounts";
          break;
        case "unverified_email":
          code = "enterpriseUnverifiedEmail";
          break;
        case "email_maxnum":
          code = "maxNumOfEmailAddressesReached";
          break;
        default:
          code = "somethingWentWrong";
      }

      return new ThirdPartyError(code, new Error(errorDescription));
    }
  }
}
