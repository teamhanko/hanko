import { ConfigClient } from "./lib/client/ConfigClient";
import { PasscodeClient } from "./lib/client/PasscodeClient";
import { PasswordClient } from "./lib/client/PasswordClient";
import { UserClient } from "./lib/client/UserClient";
import { WebauthnClient } from "./lib/client/WebauthnClient";

/**
 * A class that bundles all available SDK functions.
 *
 * @param {string} api - The URL of your Hanko API instance
 * @param {number=} timeout - The request timeout in milliseconds
 */
class Hanko {
  config: ConfigClient;
  user: UserClient;
  webauthn: WebauthnClient;
  password: PasswordClient;
  passcode: PasscodeClient;

  // eslint-disable-next-line require-jsdoc
  constructor(api: string, timeout = 13000) {
    /**
     *  @public
     *  @type {ConfigClient}
     */
    this.config = new ConfigClient(api, timeout);
    /**
     *  @public
     *  @type {UserClient}
     */
    this.user = new UserClient(api, timeout);
    /**
     *  @public
     *  @type {WebauthnClient}
     */
    this.webauthn = new WebauthnClient(api, timeout);
    /**
     *  @public
     *  @type {PasswordClient}
     */
    this.password = new PasswordClient(api, timeout);
    /**
     *  @public
     *  @type {PasscodeClient}
     */
    this.passcode = new PasscodeClient(api, timeout);
  }
}

export { Hanko };
