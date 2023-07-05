import { HttpClient, Options } from "./HttpClient";

/**
 * A class to be extended by the other client classes.
 *
 * @abstract
 * @category SDK
 * @subcategory Internal
 * @param {string} api - The URL of your Hanko API instance
 * @param {Options} options - The options that can be used
 */
abstract class Client {
  client: HttpClient;

  // eslint-disable-next-line require-jsdoc
  constructor(api: string, options: Options) {
    /**
     *  @public
     *  @type {HttpClient}
     */
    this.client = new HttpClient(api, options);
  }
}

export { Client };
