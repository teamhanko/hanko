import { Credential } from "./Dto";

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {string[]?} credentials - A list of known credential IDs on the current browser.
 */
interface LocalStorageWebauthn {
  credentials?: string[];
}

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {string=} id - The UUID of the active passcode.
 * @property {number=} ttl - Timestamp until when the passcode is valid in seconds (since January 1, 1970 00:00:00 UTC).
 * @property {number=} resendAfter - Seconds until a passcode can be resent.
 */
interface LocalStoragePasscode {
  id?: string;
  ttl?: number;
  resendAfter?: number;
}

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {number=} retryAfter - Timestamp until when the next password login can be attempted in seconds (since January 1, 1970 00:00:00 UTC).
 */
interface LocalStoragePassword {
  retryAfter?: number;
}

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {LocalStorageWebauthn=} webauthn -
 * @property {LocalStoragePasscode=} passcode - Information about the active passcode.
 * @property {LocalStoragePassword=} password - Information about the password login attempts.
 */
interface LocalStorageUser {
  webauthn?: LocalStorageWebauthn;
  passcode?: LocalStoragePasscode;
  password?: LocalStoragePassword;
}

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {Object.<string, LocalStorageUser>} - A dictionary for mapping users to their states.
 */
interface LocalStorageUsers {
  [userID: string]: LocalStorageUser;
}

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {LocalStorageUsers=} users - The user states.
 */
interface LocalStorage {
  users?: LocalStorageUsers;
}

/**
 * A class to read and write local storage contents.
 *
 * @category SDK
 * @subcategory Internal
 */
abstract class State {
  private readonly key: string;
  private ls: LocalStorage;

  public constructor(key = "hanko") {
    /**
     *  @private
     *  @type {string}
     */
    this.key = key;
    /**
     *  @private
     *  @type {LocalStorage}
     */
    this.ls = {};
  }

  /**
   * Reads and decodes the locally stored data.
   *
   * @return {LocalStorage}
   */
  read(): State {
    let store: LocalStorage;

    try {
      const data = localStorage.getItem(this.key);
      const decoded = decodeURIComponent(decodeURI(window.atob(data)));

      store = JSON.parse(decoded);
    } catch (_) {
      this.ls = {};

      return this;
    }

    this.ls = store || {};

    return this;
  }

  /**
   * Encodes and writes the data to the local storage.
   *
   * @return {void}
   */
  write(): void {
    const data = JSON.stringify(this.ls);
    const encoded = window.btoa(encodeURI(encodeURIComponent(data)));

    localStorage.setItem(this.key, encoded);
  }

  /**
   * Gets the state of the specified user.
   *
   * @protected
   * @param {string} userID - The UUID of the user.
   * @return {LocalStorageUser}
   */
  protected getUserState(userID: string): LocalStorageUser {
    this.ls.users ||= {};

    if (!Object.prototype.hasOwnProperty.call(this.ls.users, userID)) {
      this.ls.users[userID] = {};
    }

    return this.ls.users[userID];
  }

  /**
   * Converts a timestamp into remaining seconds that you can count down.
   *
   * @static
   * @protected
   * @param {number} time - Timestamp in seconds (since January 1, 1970 00:00:00 UTC).
   * @return {number}
   */
  protected static timeToRemainingSeconds(time: number = 0) {
    return time - Math.floor(Date.now() / 1000);
  }

  /**
   * Converts a number of seconds into a timestamp.
   *
   * @static
   * @protected
   * @param {number} seconds - Remaining seconds to be converted into a timestamp.
   * @return {number}
   */
  protected static remainingSecondsToTime(seconds: number = 0) {
    return Math.floor(Date.now() / 1000) + seconds;
  }
}

/**
 * A class that manages WebAuthN credentials via local storage.
 *
 * @category SDK
 * @subcategory Internal
 */
class WebauthnState extends State {
  /**
   * Gets the WebAuthN state.
   *
   * @private
   * @param {string} userID - The UUID of the user.
   * @return {LocalStorageWebauthn}
   */
  private getState(userID: string): LocalStorageWebauthn {
    return (super.getUserState(userID).webauthn ||= {});
  }

  /**
   * Reads the current states.
   *
   * @private
   * @return {WebauthnState}
   */
  read(): WebauthnState {
    super.read();

    return this;
  }

  /**
   * Gets the list of known credentials on the current browser.
   *
   * @param {string} userID - The UUID of the user.
   * @return {string[]}
   */
  getCredentials(userID: string): string[] {
    return (this.getState(userID).credentials ||= []);
  }

  /**
   * Adds the credential to the list of known credentials.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} credentialID - The WebAuthN credential ID.
   * @return {WebauthnState}
   */
  addCredential(userID: string, credentialID: string): WebauthnState {
    this.getCredentials(userID).push(credentialID);

    return this;
  }

  /**
   * Returns the intersection between the specified list of credentials and the known credentials stored in
   * the local storage.
   *
   * @param {string} userID - The UUID of the user.
   * @param {Credential[]} match - A list of credential IDs to be matched against the local storage.
   * @return {Credential[]}
   */
  matchCredentials(userID: string, match: Credential[]): Credential[] {
    return this.getCredentials(userID)
      .filter((id) => match.find((c) => c.id === id))
      .map((id: string) => ({ id } as Credential));
  }
}

/**
 * A class that manages passcodes via local storage.
 *
 * @category SDK
 * @subcategory Internal
 */
class PasscodeState extends State {
  /**
   * Get the passcode state.
   *
   * @private
   * @param {string} userID - The UUID of the user.
   * @return {LocalStoragePasscode}
   */
  private getState(userID: string): LocalStoragePasscode {
    return (super.getUserState(userID).passcode ||= {});
  }

  /**
   * Reads the current states.
   *
   * @private
   * @return {PasscodeState}
   */
  read(): PasscodeState {
    super.read();

    return this;
  }

  /**
   * Gets the UUID of the active passcode from the local storage.
   *
   * @param {string} userID - The UUID of the user.
   * @return {string}
   */
  getActiveID(userID: string): string {
    return this.getState(userID).id;
  }

  /**
   * Stores the UUID of the active passcode to the local storage.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} passcodeID - The UUID of the passcode to be set as active.
   * @return {PasscodeState}
   */
  setActiveID(userID: string, passcodeID: string): PasscodeState {
    this.getState(userID).id = passcodeID;

    return this;
  }

  /**
   * Removes the active passcode from the local storage.
   *
   * @param {string} userID - The UUID of the user.
   * @return {PasscodeState}
   */
  reset(userID: string): PasscodeState {
    const passcode = this.getState(userID);

    delete passcode.id;
    delete passcode.ttl;
    delete passcode.resendAfter;

    return this;
  }

  /**
   * Gets the TTL in seconds. When the seconds expire, the code is invalid.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getTTL(userID: string): number {
    return State.timeToRemainingSeconds(this.getState(userID).ttl);
  }

  /**
   * Sets the passcode's TTL and stores it to the local storage.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} seconds - Number of seconds the passcode is valid.
   * @return {PasscodeState}
   */
  setTTL(userID: string, seconds: number): PasscodeState {
    this.getState(userID).ttl = State.remainingSecondsToTime(seconds);

    return this;
  }

  /**
   * Gets the number of seconds until when the next passcode can be sent.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getResendAfter(userID: string): number {
    return State.timeToRemainingSeconds(this.getState(userID).resendAfter);
  }

  /**
   * Sets the number of seconds until a new passcode can be sent.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} seconds - Number of seconds the passcode is valid.
   * @return {PasscodeState}
   */
  setResendAfter(userID: string, seconds: number): PasscodeState {
    this.getState(userID).resendAfter = State.remainingSecondsToTime(seconds);

    return this;
  }
}

/**
 * A class that manages the password login state.
 *
 * @category SDK
 * @subcategory Internal
 */
class PasswordState extends State {
  /**
   * Get the password state.
   *
   * @private
   * @param {string} userID - The UUID of the user.
   * @return {LocalStoragePassword}
   */
  private getState(userID: string): LocalStoragePassword {
    return (super.getUserState(userID).password ||= {});
  }

  /**
   * Reads the current states.
   *
   * @private
   * @return {PasswordState}
   */
  read(): PasswordState {
    super.read();

    return this;
  }

  /**
   * Gets the number of seconds until when a new password login can be attempted.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getRetryAfter(userID: string): number {
    return State.timeToRemainingSeconds(this.getState(userID).retryAfter);
  }

  /**
   * Sets the number of seconds until a new password login can be attempted.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} seconds - Number of seconds the passcode is valid.
   * @return {PasswordState}
   */
  setRetryAfter(userID: string, seconds: number): PasswordState {
    this.getState(userID).retryAfter = State.remainingSecondsToTime(seconds);

    return this;
  }
}

export { WebauthnState, PasscodeState, PasswordState };
