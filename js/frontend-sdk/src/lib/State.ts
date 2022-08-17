import { Credential } from "./Dto";

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {string} id - The UUID of the active passcode.
 * @property {number} ttl - Timestamp until when the passcode is valid in seconds (since January 1, 1970 00:00:00 UTC).
 * @property {number} resendAfter - Seconds until a passcode can be resent.
 */
interface LocalStoragePasscode {
  id: string;
  ttl: number;
  resendAfter: number;
}

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {number} retryAfter - Timestamp until when the next password login can be attempted in seconds (since January 1, 1970 00:00:00 UTC).
 */
interface LocalStoragePassword {
  retryAfter: number;
}

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {string[]} webAuthnCredentials - A list of credential IDs known on the current browser.
 * @property {LocalStoragePasscode} passcode - Information about the active passcode.
 * @property {LocalStoragePassword} password - Information about the password login attempts.
 */
interface LocalStorageUser {
  webAuthnCredentials: string[];
  passcode: LocalStoragePasscode;
  password: LocalStoragePassword;
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
 * @property {LocalStorageUsers} users - The user states.
 */
interface LocalStorage {
  users?: LocalStorageUsers;
}

const initialUserState: LocalStorageUser = {
  webAuthnCredentials: [],
  passcode: { id: "", ttl: 0, resendAfter: 0 },
  password: { retryAfter: 0 },
};

/**
 * A class to read and write local storage contents.
 *
 * @category SDK
 * @subcategory Internal
 */
abstract class State {
  private key: string;

  public constructor(key = "hanko") {
    /**
     *  @private
     *  @type {string}
     */
    this.key = key;
  }

  /**
   * Reads and decodes the locally stored data.
   *
   * @return {LocalStorage}
   */
  read(): LocalStorage {
    let store: LocalStorage;
    try {
      const data = localStorage.getItem(this.key);
      const decoded = decodeURIComponent(decodeURI(window.atob(data)));

      store = JSON.parse(decoded);
    } catch (_) {
      return { users: {} } as LocalStorage;
    }

    return store;
  }

  /**
   * Encodes and writes the data to the local storage.
   *
   * @param {LocalStorage} store - The contents to be stored.
   * @return {void}
   */
  write(store: LocalStorage): void {
    const data = JSON.stringify(store);
    const encoded = window.btoa(encodeURI(encodeURIComponent(data)));

    localStorage.setItem(this.key, encoded);
  }

  /**
   * Gets the state of the specified user.
   *
   * @param {string} userID - The UUID of the user.
   * @return {LocalStorageUser}
   */
  getUserState(userID: string) {
    const store = this.read();
    const exists = Object.prototype.hasOwnProperty.call(store.users, userID);

    return exists ? store.users[userID] : initialUserState;
  }

  /**
   * Sets the state of the specified user.
   *
   * @param {string} userID - The UUID of the user.
   * @param {LocalStorageUser} state - The state of the specified user.
   * @return {LocalStorageUser}
   */
  setUserState(userID: string, state: LocalStorageUser) {
    const store = this.read();

    store.users[userID] = state;
    this.write(store);
  }

  /**
   * Converts a timestamp into remaining seconds that you can count down.
   *
   * @param {number} time - Timestamp in seconds (since January 1, 1970 00:00:00 UTC).
   * @return {number}
   */
  timeToRemainingSeconds(time: number = 0) {
    return time - Math.floor(Date.now() / 1000);
  }

  /**
   * Converts a number of seconds into a timestamp.
   *
   * @param {number} seconds - Remaining seconds to be converted into a timestamp.
   * @return {number}
   */
  remainingSecondsToTime(seconds: number = 0) {
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
   * Adds the given 'credentialID' to the list of known credentials and stores the updated list to the local storage.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} credentialID - The credential ID to be saved.
   * @return {void}
   */
  addCredentialID(userID: string, credentialID: string): void {
    const state = super.getUserState(userID);

    state.webAuthnCredentials.push(credentialID);
    this.setUserState(userID, state);
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
    const { webAuthnCredentials } = super.getUserState(userID);
    return webAuthnCredentials
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
   * Gets the UUID of the active passcode from the local storage.
   *
   * @param {string} userID - The UUID of the user.
   * @return {string}
   */
  getActiveID(userID: string): string {
    const { passcode } = this.getUserState(userID);

    return passcode.id;
  }

  /**
   * Stores the UUID of the active passcode to the local storage.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} passcodeID - The UUID of the passcode to be set as active.
   * @return {void}
   */
  setActiveID(userID: string, passcodeID: string) {
    const state = this.getUserState(userID);

    state.passcode.id = passcodeID;
    this.setUserState(userID, state);
  }

  /**
   * Removes the active passcode from the local storage.
   *
   * @param {string} userID - The UUID of the user.
   * @return {void}
   */
  removeActive(userID: string) {
    const state = this.getUserState(userID);

    state.passcode.id = initialUserState.passcode.id;
    state.passcode.ttl = initialUserState.passcode.ttl;
    this.setUserState(userID, state);
  }

  /**
   * Gets the TTL in seconds. When the seconds expire, the code is invalid.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getTTL(userID: string): number {
    const state = this.getUserState(userID);

    return this.timeToRemainingSeconds(state.passcode.ttl);
  }

  /**
   * Sets the passcode's TTL and stores it to the local storage.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} seconds - Number of seconds the passcode is valid.
   * @return {void}
   */
  setTTL(userID: string, seconds: number): void {
    const state = this.getUserState(userID);

    state.passcode.ttl = this.remainingSecondsToTime(seconds);
    this.setUserState(userID, state);
  }

  /**
   * Gets the number of seconds until when the next passcode can be sent.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getResendAfter(userID: string): number {
    const { passcode } = this.getUserState(userID);

    return this.timeToRemainingSeconds(passcode.resendAfter);
  }

  /**
   * Sets the number of seconds until a new passcode can be sent.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} seconds - Number of seconds the passcode is valid.
   * @return {void}
   */
  setResendAfter(userID: string, seconds: number): void {
    const state = this.getUserState(userID);

    state.passcode.resendAfter = this.remainingSecondsToTime(seconds);
    this.setUserState(userID, state);
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
   * Gets the number of seconds until when a new password login can be attempted.
   *
   * @param {string} userID - The UUID of the user.
   * @return {number}
   */
  getRetryAfter(userID: string): number {
    const state = this.getUserState(userID);

    return this.timeToRemainingSeconds(state.password.retryAfter);
  }

  /**
   * Sets the number of seconds until a new password login can be attempted.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} seconds - Number of seconds the passcode is valid.
   * @return {void}
   */
  setRetryAfter(userID: string, seconds: number): void {
    const state = this.getUserState(userID);

    state.password.retryAfter = this.remainingSecondsToTime(seconds);
    this.setUserState(userID, state);
  }
}

export { WebauthnState, PasscodeState, PasswordState };
