import { LocalStorageUsers } from "./UserState";

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
 * @abstract
 * @param {string} key - The local storage key.
 * @category SDK
 * @subcategory Internal
 */
abstract class State {
  private readonly key: string;
  protected ls: LocalStorage;

  // eslint-disable-next-line require-jsdoc
  constructor(key = "hanko") {
    /**
     *  @private
     *  @type {string}
     */
    this.key = key;
    /**
     *  @protected
     *  @type {LocalStorage}
     */
    this.ls = {};
  }

  /**
   * Reads and decodes the locally stored data.
   *
   * @protected
   * @return {State}
   */
  protected read(): State {
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
   * @return {State}
   */
  write(): State {
    const data = JSON.stringify(this.ls);
    const encoded = window.btoa(encodeURI(encodeURIComponent(data)));

    localStorage.setItem(this.key, encoded);

    return this;
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

export { State };
