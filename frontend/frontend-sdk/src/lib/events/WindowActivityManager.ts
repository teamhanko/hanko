// Callback type for handling window activity changes.
type Callback = () => void;

/**
 * Manages window focus and blur events.
 *
 * @class
 * @category SDK
 * @subcategory Internal
 * @param {Callback} onActivityCallback - Callback to invoke when the window gains focus.
 * @param {Callback} onInactivityCallback - Callback to invoke when the window loses focus.
 */
export class WindowActivityManager {
  private readonly onActivityCallback: Callback; // Callback for when the window or tab gains focus.
  private readonly onInactivityCallback: Callback; // Callback for when the window or tab loses focus.

  // eslint-disable-next-line require-jsdoc
  constructor(onActivityCallback: Callback, onInactivityCallback: Callback) {
    this.onActivityCallback = onActivityCallback;
    this.onInactivityCallback = onInactivityCallback;

    // Attach event listeners for focus and blur
    window.addEventListener("focus", this.handleFocus);
    window.addEventListener("blur", this.handleBlur);
    document.addEventListener("visibilitychange", this.handleVisibilityChange);
  }

  /**
   * Handles the focus event and invokes the activity callback.
   * @private
   */
  private handleFocus = (): void => {
    this.onActivityCallback();
  };

  /**
   * Handles the blur event and invokes the inactivity callback.
   * @private
   */
  private handleBlur = (): void => {
    this.onInactivityCallback();
  };

  /**
   * Handles the visibility change event and invokes appropriate callbacks.
   * @private
   */
  private handleVisibilityChange = (): void => {
    if (document.visibilityState === "visible") {
      this.onActivityCallback();
    } else {
      this.onInactivityCallback();
    }
  };

  /**
   * Checks if the current window has focus.
   * @returns {boolean} True if the window has focus; otherwise, false.
   */
  hasFocus = (): boolean => {
    return document.hasFocus();
  };
}
