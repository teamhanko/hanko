import { Client } from "./Client";
import {
  EmailAddressAlreadyExistsError,
  MaxNumOfEmailAddressesReachedError,
  TechnicalError,
  UnauthorizedError,
} from "../Errors";
import { Email, Emails } from "../Dto";

/**
 * Manages email addresses of the current user.
 *
 * @constructor
 * @category SDK
 * @subcategory Clients
 * @extends {Client}
 */
class EmailClient extends Client {
  /**
   * Returns a list of all email addresses assigned to the current user.
   *
   * @return {Promise<Emails>}
   * @throws {UnauthorizedError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @see https://docs.hanko.io/api/public#tag/Email-Management/operation/listEmails
   */
  async list(): Promise<Emails> {
    const response = await this.client.get("/emails");

    if (response.status === 401) {
      this.client.dispatcher.dispatchSessionExpiredEvent();
      throw new UnauthorizedError();
    } else if (!response.ok) {
      throw new TechnicalError();
    }

    return response.json();
  }

  /**
   * Adds a new email address to the current user.
   *
   * @param {string} address - The email address to be added.
   * @return {Promise<Email>}
   * @throws {EmailAddressAlreadyExistsError}
   * @throws {MaxNumOfEmailAddressesReachedError}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @throws {UnauthorizedError}
   * @see https://docs.hanko.io/api/public#tag/Email-Management/operation/createEmail
   */
  async create(address: string): Promise<Email> {
    const response = await this.client.post("/emails", { address });

    if (response.ok) {
      return response.json();
    }

    if (response.status === 400) {
      throw new EmailAddressAlreadyExistsError();
    } else if (response.status === 401) {
      this.client.dispatcher.dispatchSessionExpiredEvent();
      throw new UnauthorizedError();
    } else if (response.status === 409) {
      throw new MaxNumOfEmailAddressesReachedError();
    }

    throw new TechnicalError();
  }

  /**
   * Marks the specified email address as primary.
   *
   * @param {string} emailID - The ID of the email address to be updated
   * @return {Promise<void>}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @throws {UnauthorizedError}
   * @see https://docs.hanko.io/api/public#tag/Email-Management/operation/setPrimaryEmail
   */
  async setPrimaryEmail(emailID: string): Promise<void> {
    const response = await this.client.post(`/emails/${emailID}/set_primary`);

    if (response.status === 401) {
      this.client.dispatcher.dispatchSessionExpiredEvent();
      throw new UnauthorizedError();
    } else if (!response.ok) {
      throw new TechnicalError();
    }

    return;
  }

  /**
   * Deletes the specified email address.
   *
   * @param {string} emailID - The ID of the email address to be deleted
   * @return {Promise<void>}
   * @throws {RequestTimeoutError}
   * @throws {TechnicalError}
   * @throws {UnauthorizedError}
   * @see https://docs.hanko.io/api/public#tag/Email-Management/operation/deleteEmail
   */
  async delete(emailID: string): Promise<void> {
    const response = await this.client.delete(`/emails/${emailID}`);

    if (response.status === 401) {
      this.client.dispatcher.dispatchSessionExpiredEvent();
      throw new UnauthorizedError();
    } else if (!response.ok) {
      throw new TechnicalError();
    }

    return;
  }
}

export { EmailClient };
