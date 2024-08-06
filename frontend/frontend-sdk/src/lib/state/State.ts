import { LocalStorageSession } from "./session/SessionState";

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {LocalStorageUsers=} users - The user states.
 */
interface LocalStorage {
  session?: LocalStorageSession;
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
  key: string;
  ls: LocalStorage;

  // eslint-disable-next-line require-jsdoc
  constructor(key: string) {
    /**
     *  @private
     *  @type {string}
     */
    this.key = key;
    /**
     *  @type {LocalStorage}
     */
    this.ls = {};
  }

  /**
   * Reads and decodes the locally stored data.
   *
   * @return {State}
   */
  read(): State {
    let store: LocalStorage;

    try {
      const data = localStorage.getItem(this.key);
      const decoded = decodeURIComponent(decodeURI(window.atob(data)));
      store = JSON.parse(decoded);
    } catch (e) {
      this.ls = {};

      return this;
    }

    this.ls = store;

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
   * @param {number} time - Timestamp in seconds (since January 1, 1970 00:00:00 UTC).
   * @return {number}
   */
  static timeToRemainingSeconds(time = 0) {
    return time - Math.floor(Date.now() / 1000);
  }

  /**
   * Converts a number of seconds into a timestamp.
   *
   * @static
   * @param {number} seconds - Remaining seconds to be converted into a timestamp.
   * @return {number}
   */
  static remainingSecondsToTime(seconds = 0) {
    return Math.floor(Date.now() / 1000) + seconds;
  }
}

export { State };
export type { LocalStorage };
