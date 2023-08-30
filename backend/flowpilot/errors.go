package flowpilot

import (
	"fmt"
	"net/http"
)

// flowpilotError defines the interface for custom error types in the Flowpilot package.
type flowpilotError interface {
	error

	Unwrap() error
	Code() string
	Message() string

	toPublicError(debug bool) PublicError
}

// FlowError is an interface representing flow-related errors.
type FlowError interface {
	flowpilotError

	Wrap(error) FlowError
	Status() int
}

// InputError is an interface representing input-related errors.
type InputError interface {
	flowpilotError

	Wrap(error) InputError
}

// defaultError is a base struct for custom error types.
type defaultError struct {
	origin    error  // The error origin.
	code      string // Unique error code.
	message   string // Contains a description of the error.
	errorText string // The string representation of the error.
}

// Code returns the error code.
func (e *defaultError) Code() string {
	return e.code
}

// Message returns the error message.
func (e *defaultError) Message() string {
	return e.message
}

// Unwrap returns the wrapped error.
func (e *defaultError) Unwrap() error {
	return e.origin
}

// Error returns the formatted error message.
func (e *defaultError) Error() string {
	return e.errorText
}

// toPublicError converts the error to a PublicError for public exposure.
func (e *defaultError) toPublicError(debug bool) PublicError {
	pe := PublicError{
		Code:    e.Code(),
		Message: e.Message(),
	}

	if debug && e.origin != nil {
		str := e.origin.Error()
		pe.Origin = &str
	}

	return pe
}

// defaultFlowError is a struct for flow-related errors.
type defaultFlowError struct {
	defaultError

	status int // The suggested HTTP status code.
}

func createErrorText(code, message string, origin error) string {
	txt := fmt.Sprintf("%s - %s", code, message)
	if origin != nil {
		txt = fmt.Sprintf("%s: %s", txt, origin.Error())
	}
	return txt
}

// NewFlowError creates a new FlowError instance.
func NewFlowError(code, message string, status int) FlowError {
	return newFlowErrorWithOrigin(code, message, status, nil)
}

// newFlowErrorWithOrigin creates a new FlowError instance with an origin error.
func newFlowErrorWithOrigin(code, message string, status int, origin error) FlowError {
	e := defaultError{
		origin:    origin,
		code:      code,
		message:   message,
		errorText: createErrorText(code, message, origin),
	}

	return &defaultFlowError{defaultError: e, status: status}
}

// Status returns the suggested HTTP status code.
func (e *defaultFlowError) Status() int {
	return e.status
}

// Wrap wraps the error with another error.
func (e *defaultFlowError) Wrap(err error) FlowError {
	return newFlowErrorWithOrigin(e.code, e.message, e.status, err)
}

// defaultInputError is a struct for input-related errors.
type defaultInputError struct {
	defaultError
}

// NewInputError creates a new InputError instance.
func NewInputError(code, message string) InputError {
	return newInputErrorWithOrigin(code, message, nil)
}

// newInputErrorWithOrigin creates a new InputError instance with an origin error.
func newInputErrorWithOrigin(code, message string, origin error) InputError {
	e := defaultError{
		origin:    origin,
		code:      code,
		message:   message,
		errorText: createErrorText(code, message, origin),
	}

	return &defaultInputError{defaultError: e}
}

// Wrap wraps the error with another error.
func (e *defaultInputError) Wrap(err error) InputError {
	return newInputErrorWithOrigin(e.code, e.message, err)
}

// Predefined flow error types
var (
	ErrorTechnical             = NewFlowError("technical_error", "Something went wrong.", http.StatusInternalServerError)
	ErrorFlowExpired           = NewFlowError("flow_expired_error", "The flow has expired.", http.StatusGone)
	ErrorFlowDiscontinuity     = NewFlowError("flow_discontinuity_error", "The flow can't be continued.", http.StatusInternalServerError)
	ErrorOperationNotPermitted = NewFlowError("operation_not_permitted_error", "The operation is not permitted.", http.StatusForbidden)
	ErrorFormDataInvalid       = NewFlowError("form_data_invalid_error", "Form data invalid.", http.StatusBadRequest)
	ErrorActionParamInvalid    = NewFlowError("action_param_invalid_error", "Action parameter is invalid.", http.StatusBadRequest)
)

// Predefined input error types
var (
	ErrorEmailInvalid  = NewInputError("email_invalid_error", "The email address is invalid.")
	ErrorValueMissing  = NewInputError("value_missing_error", "Missing value.")
	ErrorValueInvalid  = NewInputError("value_invalid_error", "The value is invalid.")
	ErrorValueTooLong  = NewInputError("value_too_long_error", "Value is too long.")
	ErrorValueTooShort = NewInputError("value_too_short_error", "Value is too short.")
)
