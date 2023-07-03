import { HttpClient } from "./HttpClient";
import { Options } from "../../Hanko";

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
  constructor(api: string, options: Options) {
    /**
     *  @public
     *  @type {HttpClient}
     */
    this.client = new HttpClient(api, options);
  }
}

export { Client };
