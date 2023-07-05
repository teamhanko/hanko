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
 * The options for the Hanko class
 *
 * @interface
 * @property {number=} timeout - The http request timeout in milliseconds. Defaults to 13000ms
 * @property {string=} cookieName - The name of the session cookie set from the SDK. Defaults to "hanko"
 * @property {string=} storageKey - The prefix / name of the local storage keys. Defaults to "hanko"
 */
export interface Options {
  timeout?: number;
  cookieName?: string;
  storageKey?: string;
}

/**
 * A class that bundles all available SDK functions.
 *
 * @extends {Listener}
 * @param {string} api - The URL of your Hanko API instance
 * @param {Options=} options - The options that can be used
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
  constructor(api: string, options?: Options) {
    super();
    const opts: InternalOptions = {
      timeout: 13000,
      cookieName: "hanko",
      storageKey: "hanko",
    };
    if (options?.cookieName !== undefined) {
      opts.cookieName = options.cookieName;
    }
    if (options?.timeout !== undefined) {
      opts.timeout = options.timeout;
    }
    if (options?.storageKey !== undefined) {
      opts.storageKey = options.storageKey;
    }

    this.api = api;
    /**
     *  @public
     *  @type {ConfigClient}
     */
    this.config = new ConfigClient(api, opts);
    /**
     *  @public
     *  @type {UserClient}
     */
    this.user = new UserClient(api, opts);
    /**
     *  @public
     *  @type {WebauthnClient}
     */
    this.webauthn = new WebauthnClient(api, opts);
    /**
     *  @public
     *  @type {PasswordClient}
     */
    this.password = new PasswordClient(api, opts);
    /**
     *  @public
     *  @type {PasscodeClient}
     */
    this.passcode = new PasscodeClient(api, opts);
    /**
     *  @public
     *  @type {EmailClient}
     */
    this.email = new EmailClient(api, opts);
    /**
     *  @public
     *  @type {ThirdPartyClient}
     */
    this.thirdParty = new ThirdPartyClient(api, opts);
    /**
     *  @public
     *  @type {TokenClient}
     */
    this.token = new TokenClient(api, opts);
    /**
     *  @public
     *  @type {Relay}
     */
    this.relay = new Relay({ ...opts });
    /**
     *  @public
     *  @type {Session}
     */
    this.session = new Session({ ...opts });
  }
}

// eslint-disable-next-line require-jsdoc
export interface InternalOptions {
  timeout: number;
  cookieName: string;
  storageKey: string;
}

export { Hanko };
