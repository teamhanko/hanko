package flowpilot

// TODO: Guess it would be nice to add an error interface

// ErrorType represents a custom error type with a code and message.
type ErrorType struct {
	Code    string `json:"code"`    // Unique error code.
	Message string `json:"message"` // Description of the error.
}

// Predefined error types
var (
	TechnicalError             = &ErrorType{Code: "technical_error", Message: "Something went wrong."}
	FlowExpiredError           = &ErrorType{Code: "flow_expired_error", Message: "The flow has expired."}
	FlowDiscontinuityError     = &ErrorType{Code: "flow_discontinuity_error", Message: "Thr flow can't be continued."}
	OperationNotPermittedError = &ErrorType{Code: "operation_not_permitted_error", Message: "The operation is not permitted."}
	FormDataInvalidError       = &ErrorType{Code: "form_data_invalid_error", Message: "Form data invalid."}
	EmailInvalidError          = &ErrorType{Code: "email_invalid_error", Message: "The email address is invalid."}
	ValueMissingError          = &ErrorType{Code: "value_missing_error", Message: "Missing value."}
	ValueInvalidError          = &ErrorType{Code: "value_invalid_error", Message: "The value is invalid."}
	ValueTooLongError          = &ErrorType{Code: "value_too_long_error", Message: "Value is too long."}
	ValueTooShortError         = &ErrorType{Code: "value_too_short_error", Message: "Value is too short."}
)
