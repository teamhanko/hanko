import { Listener } from "./lib/events/Listener";
import { Relay } from "./lib/events/Relay";
import { Cookie, CookieSameSite } from "./lib/Cookie";
import { SessionClient } from "./lib/client/SessionClient";
import { HttpClient, SessionTokenLocation } from "./lib/client/HttpClient";
import { FlowName } from "./lib/flow-api/types/flow";
import { StateCreateConfig, State } from "./lib/flow-api/State";
import { UserClient } from "./lib/client/UserClient";
import { SessionCheckResponse } from "./lib/Dto";
import { User } from "./lib/flow-api/types/payload";

/**
 * The options for the Hanko class
 *
 * @interface
 * @property {number=} timeout - The http request timeout in milliseconds. Defaults to 13000ms
 * @property {string=} cookieName - The name of the session cookie set from the SDK. Defaults to "hanko"
 * @property {string=} cookieDomain - The domain where the cookie set from the SDK is available. Defaults to the domain of the page where the cookie was created.
 * @property {string=} cookieSameSite - Specify whether/when cookies are sent with cross-site requests. Defaults to "lax".
 * @property {string=} localStorageKey - The prefix / name of the local storage keys. Defaults to "hanko"
 * @property {string=} lang - Used to convey the preferred language to the API, e.g. for translating outgoing emails.
 *                            It is transmitted to the API in a custom header (X-Language).
 *                            Should match one of the supported languages ("bn", "de", "en", "fr", "it, "nl", "pt-BR", "zh")
 *                            if email delivery by Hanko is enabled. If email delivery by Hanko is disabled and the
 *                            relying party configures a webhook for the "email.send" event, then the set language is
 *                            reflected in the payload of the token contained in the webhook request.
 * @property {number=} sessionCheckInterval -  Interval for session validity checks in milliseconds. Must be greater than 3000 (3s), defaults to 3000 otherwise.
 * @property {string=} sessionCheckChannelName - The broadcast channel name for inter-tab communication.
 * @property {string=} sessionTokenLocation - The location where the session token is stored.
 * @property {string=} tenantId - The tenant ID for multi-tenant support. Transmitted via X-Tenant-ID header.
 */
export interface HankoOptions {
  timeout?: number;
  cookieName?: string;
  cookieDomain?: string;
  cookieSameSite?: CookieSameSite;
  localStorageKey?: string;
  lang?: string;
  sessionCheckInterval?: number;
  sessionCheckChannelName?: string;
  sessionTokenLocation?: SessionTokenLocation;
  tenantId?: string;
}

/**
 * A class that bundles all available SDK functions.
 *
 * @extends {Listener}
 * @param {string} api - The URL of your Hanko API instance
 * @param {HankoOptions=} options - The options that can be used
 */
class Hanko extends Listener {
  private readonly session: SessionClient;
  private readonly user: UserClient;
  private readonly cookie: Cookie;
  public readonly client: HttpClient;
  public readonly relay: Relay;

  // eslint-disable-next-line require-jsdoc
  constructor(api: string, options?: HankoOptions) {
    super();
    const opts: HankoOptions = {
      timeout: 13000,
      cookieName: "hanko",
      localStorageKey: "hanko",
      sessionCheckInterval: 30000,
      sessionCheckChannelName: "hanko-session-check",
      ...options,
    };

    /**
     *  @public
     *  @type {Client}
     */
    this.client = new HttpClient(api, opts);
    /**
     *  @public
     *  @type {SessionClient}
     */
    this.session = new SessionClient(api, opts);
    /**
     *  @public
     *  @type {SessionClient}
     */
    this.user = new UserClient(api, opts);
    /**
     *  @public
     *  @type {Relay}
     */
    this.relay = new Relay(api, opts);
    /**
     *  @public
     *  @type {Cookie}
     */
    this.cookie = new Cookie(opts);
  }

  /**
   * Sets the preferred language on the underlying sub-clients. The clients'
   * base HttpClient uses this language to transmit an X-Language header to the
   * API which is then used to e.g. translate outgoing emails.
   *
   * @public
   * @param lang {string} - The preferred language to convey to the API.
   */
  setLang(lang: string) {
    this.client.lang = lang;
  }

  /**
   * Sets the tenant ID for multi-tenant support. The tenant ID is transmitted
   * to the API via X-Tenant-ID header.
   *
   * @public
   * @param tenantId {string} - The tenant UUID to send with requests.
   */
  setTenantId(tenantId: string) {
    this.client.tenantId = tenantId;
  }

  /**
   * Creates a new flow state for the specified flow.
   *
   * This method initializes a state by either loading from cache (if configured) or fetching from the server.
   * It uses the provided configuration to control caching, event dispatching, and auto-step behavior.
   *
   * @param {FlowName} flowName - The name of the flow to create a state for.
   * @param {StateCreateConfig} [config={}] - Configuration options for state creation.
   * @param {boolean} [config.dispatchAfterStateChangeEvent=true] - Whether to dispatch an event after the state changes.
   * @param {AutoStepExclusion} [config.excludeAutoSteps=null] - States to exclude from auto-step processing, or `"all"` to skip all auto-steps.
   * @param {string} [config.cacheKey="hanko-flow-state"] - Key used for caching the state in localStorage.
   * @param {boolean} [config.loadFromCache=true] - Whether to attempt loading the state from cache.
   * @returns {Promise<AnyState>} A promise that resolves to the created flow state.
   * @category SDK
   * @subcategory FlowAPI
   */
  createState(flowName: FlowName, config: StateCreateConfig = {}) {
    return State.create(this, flowName, config);
  }

  /**
   * Retrieves the current user's profile information.
   *
   * @public
   * @returns {Promise<User>} A promise that resolves to the user object
   * @throws {UnauthorizedError} If the user is not authenticated
   * @throws {TechnicalError} If an unexpected error occurs
   */
  async getUser(): Promise<User> {
    return this.user.getCurrent();
  }

  /**
   * Validates the current session.
   *
   * @public
   * @returns {Promise<SessionCheckResponse>} A promise that resolves to the session check response
   */
  async validateSession(): Promise<SessionCheckResponse> {
    return this.session.validate();
  }

  /**
   * Retrieves the current session token from the authentication cookie.
   *
   * @public
   * @returns {string} The session token
   */
  getSessionToken(): string {
    return this.cookie.getAuthCookie();
  }

  /**
   * Logs out the current user by invalidating the session.
   *
   * @public
   * @returns {Promise<void>} A promise that resolves when the logout is complete
   */
  async logout(): Promise<void> {
    return this.user.logout();
  }
}

export { Hanko };
