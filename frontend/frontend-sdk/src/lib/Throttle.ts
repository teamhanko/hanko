/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {boolean=} leading - Whether to allow the function to be called on the leading edge of the wait timeout.
 * @property {boolean=} trailing - Whether to allow the function to be called on the trailing edge of the wait timeout.
 */
interface ThrottleOptions {
  leading?: boolean;
  trailing?: boolean;
}

// eslint-disable-next-line no-unused-vars
type ThrottledFunction<T extends (...args: any[]) => any> = (
  // eslint-disable-next-line no-unused-vars
  ...args: Parameters<T>
) => void;

/**
 * Provides throttle functionality.
 *
 * @hideconstructor
 * @category SDK
 * @subcategory Internal
 */
export class Throttle {
  /**
   * Throttles a function, ensuring that it can only be called once per `wait` milliseconds.
   *
   * @static
   * @param {function} func - The function to throttle.
   * @param {number} wait - The number of milliseconds to wait between function invocations.
   * @param {ThrottleOptions} options - Optional configuration for the throttle.
   * @returns {function} A throttled version of the original function.
   */
  // eslint-disable-next-line no-unused-vars,require-jsdoc
  static throttle<T extends (...args: any[]) => any>(
    func: T,
    wait: number,
    options: ThrottleOptions = {}
  ): ThrottledFunction<T> {
    const { leading = true, trailing = true } = options;
    let context: any;
    let args: any;
    let timeoutID: number;
    let previous = 0;

    // This function is used to invoke the original function.
    const executeThrottledFunction = () => {
      // If 'leading' is false and this is not the first invocation of the throttled function, set 'previous' to 0 to
      // ensure that the function is not called immediately.
      previous = leading === false ? 0 : Date.now();
      timeoutID = null;
      // Invoke the original function.
      func.apply(context, args);
    };

    // This is the throttled function that will be returned.
    const throttled = function (...funcArgs: Parameters<T>) {
      const now = Date.now();

      // If this is the first time the throttled function is being called, and 'leading' is false,
      // set 'previous' to the current time to ensure that the function is not called immediately.
      if (!previous && leading === false) previous = now;

      // The remaining wait time.
      const remaining = wait - (now - previous);

      // Save the context and arguments of the function call.
      // eslint-disable-next-line no-invalid-this
      context = this;
      args = funcArgs;

      // Check whether it's time to call the function immediately based on the leading and trailing options. If leading
      // is enabled and there was no previous invocation, or if trailing is enabled and the wait time has already passed,
      // the function will be invoked immediately.
      if (remaining <= 0 || remaining > wait) {
        // If there is a pending timeout, clear it.
        if (timeoutID) {
          window.clearTimeout(timeoutID);
          timeoutID = null;
        }

        // Invoke the original function and update the previous timestamp.
        previous = now;
        func.apply(context, args);
      } else if (!timeoutID && trailing !== false) {
        // If there is no pending timeout and trailing is allowed, start a new timeout.
        timeoutID = window.setTimeout(executeThrottledFunction, remaining);
      }
    };

    return throttled;
  }
}
