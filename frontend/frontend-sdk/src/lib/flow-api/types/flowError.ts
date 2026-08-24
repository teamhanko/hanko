/**
 * An error returned by the flow API, either at the top level of a {@link State} (a general failure)
 * or attached to an individual {@link Input} (a validation error for that specific field).
 * @interface
 * @category SDK
 * @subcategory Flow Errors
 * @property {string} code - A machine-readable error code.
 * @property {string} message - A human-readable error message.
 * @property {string} [cause] - Additional detail about what caused the error, if available.
 */
export interface FlowError {
  code: string;
  message: string;
  cause?: string;
}
