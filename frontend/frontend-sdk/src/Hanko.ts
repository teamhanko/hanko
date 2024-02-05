import { ConfigClient } from "./lib/client/ConfigClient";
import { EnterpriseClient } from "./lib/client/EnterpriseClient";
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
import { CookieSameSite } from "./lib/Cookie";

/**
 * The options for the Hanko class
 *
 * @interface
 * @property {number=} timeout - The http request timeout in milliseconds. Defaults to 13000ms
 * @property {string=} cookieName - The name of the session cookie set from the SDK. Defaults to "hanko"
 * @property {string=} cookieDomain - The domain where the cookie set from the SDK is available. Defaults to the domain of the page where the cookie was created.
 * @property {string=} cookieSameSite - Specify whether/when cookies are sent with cross-site requests. Defaults to "lax".
 * @property {string=} localStorageKey - The prefix / name of the local storage keys. Defaults to "hanko"
 */
export interface HankoOptions {
  timeout?: number;
  cookieName?: string;
  cookieDomain?: string;
  cookieSameSite?: CookieSameSite;
  localStorageKey?: string;
}

/**
 * A class that bundles all available SDK functions.
 *
 * @extends {Listener}
 * @param {string} api - The URL of your Hanko API instance
 * @param {HankoOptions=} options - The options that can be used
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
  enterprise: EnterpriseClient;
  token: TokenClient;
  relay: Relay;
  session: Session;

  // eslint-disable-next-line require-jsdoc
  constructor(api: string, options?: HankoOptions) {
    super();
    const opts: InternalOptions = {
      timeout: 13000,
      cookieName: "hanko",
      localStorageKey: "hanko",
    };
    if (options?.cookieName !== undefined) {
      opts.cookieName = options.cookieName;
    }
    if (options?.timeout !== undefined) {
      opts.timeout = options.timeout;
    }
    if (options?.localStorageKey !== undefined) {
      opts.localStorageKey = options.localStorageKey;
    }
    if (options?.cookieDomain !== undefined) {
      opts.cookieDomain = options.cookieDomain;
    }
    if (options?.cookieSameSite !== undefined) {
      opts.cookieSameSite = options.cookieSameSite;
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
     *  @type {EnterpriseClient}
     */
    this.enterprise = new EnterpriseClient(api, opts);
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
  cookieDomain?: string;
  cookieSameSite?: CookieSameSite;
  localStorageKey: string;
}

export { Hanko };
