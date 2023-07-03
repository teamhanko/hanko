import { ConfigClient } from "./lib/client/ConfigClient";
import { PasscodeClient } from "./lib/client/PasscodeClient";
import { PasswordClient } from "./lib/client/PasswordClient";
import { UserClient } from "./lib/client/UserClient";
import { WebauthnClient } from "./lib/client/WebauthnClient";
import { EmailClient } from "./lib/client/EmailClient";
import { ThirdPartyClient } from "./lib/client/ThirdPartyClient";
import { TokenClient } from "./lib/client/TokenClient";
import { Listener } from "./lib/events/Listener";
import { Relay } from "./lib/events/Relay";
import { Session } from "./lib/Session";

/**
 * A class that bundles all available SDK functions.
 *
 * @extends {Listener}
 * @param {string} api - The URL of your Hanko API instance
 * @param {number=} timeout - The request timeout in milliseconds
 */
class Hanko extends Listener {
  api: string;
  config: ConfigClient;
  user: UserClient;
  webauthn: WebauthnClient;
  password: PasswordClient;
  passcode: PasscodeClient;
  email: EmailClient;
  thirdParty: ThirdPartyClient;
  token: TokenClient;
  relay: Relay;
  session: Session;

  // eslint-disable-next-line require-jsdoc
  constructor(
    api: string,
    options: Options = { timeout: 13000, cookieName: "hanko" }
  ) {
    super();
    if (options.cookieName === undefined) {
      options.cookieName = "hanko";
    }
    if (options.timeout === undefined) {
      options.timeout = 13000;
    }

    this.api = api;
    /**
     *  @public
     *  @type {ConfigClient}
     */
    this.config = new ConfigClient(api, options);
    /**
     *  @public
     *  @type {UserClient}
     */
    this.user = new UserClient(api, options);
    /**
     *  @public
     *  @type {WebauthnClient}
     */
    this.webauthn = new WebauthnClient(api, options);
    /**
     *  @public
     *  @type {PasswordClient}
     */
    this.password = new PasswordClient(api, options);
    /**
     *  @public
     *  @type {PasscodeClient}
     */
    this.passcode = new PasscodeClient(api, options);
    /**
     *  @public
     *  @type {EmailClient}
     */
    this.email = new EmailClient(api, options);
    /**
     *  @public
     *  @type {ThirdPartyClient}
     */
    this.thirdParty = new ThirdPartyClient(api, options);
    /**
     *  @public
     *  @type {TokenClient}
     */
    this.token = new TokenClient(api, options);
    /**
     *  @public
     *  @type {Relay}
     */
    this.relay = new Relay(options.cookieName);
    /**
     *  @public
     *  @type {Session}
     */
    this.session = new Session(options.cookieName);
  }
}

export interface Options {
  timeout?: number;
  cookieName?: string;
}

export { Hanko };
