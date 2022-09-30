import { HttpClient } from "./HttpClient";

/**
 * A class to be extended by the other client classes.
 *
 * @abstract
 * @category SDK
 * @subcategory Internal
 * @param {string} api - The URL of your Hanko API instance
 * @param {number=} timeout - The request timeout in milliseconds
 */
abstract class Client {
  client: HttpClient;

  // eslint-disable-next-line require-jsdoc
  constructor(api: string, timeout = 13000) {
    /**
     *  @protected
     *  @type {HttpClient}
     */
    this.client = new HttpClient(api, timeout);
  }
}

export { Client };
