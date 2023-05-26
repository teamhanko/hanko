/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {Object.<string, LocalStorageUser>} - A dictionary for mapping users to their states.
 */
import { State } from "../State";

import { LocalStorageWebauthn } from "./WebauthnState";
import { LocalStoragePasscode } from "./PasscodeState";
import { LocalStoragePassword } from "./PasswordState";

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {LocalStorageWebauthn=} webauthn - Information about WebAuthn credentials.
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
export interface LocalStorageUsers {
  [userID: string]: LocalStorageUser;
}

/**
 * A class to read and write local storage contents.
 *
 * @abstract
 * @extends State
 * @param {string} key - The local storage key.
 * @category SDK
 * @subcategory Internal
 */
abstract class UserState extends State {
  // eslint-disable-next-line require-jsdoc
  constructor() {
    super("hanko");
  }

  /**
   * Gets the state of the specified user.
   *
   * @param {string} userID - The UUID of the user.
   * @return {LocalStorageUser}
   */
  getUserState(userID: string): LocalStorageUser {
    this.ls.users ||= {};

    if (!Object.prototype.hasOwnProperty.call(this.ls.users, userID)) {
      this.ls.users[userID] = {};
    }

    return this.ls.users[userID];
  }
}

export { UserState };
