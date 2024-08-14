package flowpilot

import (
	"fmt"
	"strings"
)

// flowpilotError defines the interface for custom error types in the Flowpilot package.
type flowpilotError interface {
	error

	Unwrap() error
	Code() string
	Message() string

	toResponseError(debug bool) *ResponseError
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
	cause     error  // The error cause.
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
	return e.cause
}

// Error returns the formatted error message.
func (e *defaultError) Error() string {
	return e.errorText
}

// toResponseError converts the error to a ResponseError for public exposure.
func (e *defaultError) toResponseError(debug bool) *ResponseError {
	publicError := &ResponseError{
		Code:    e.Code(),
		Message: e.Message(),
	}

	if e.cause != nil {
		cause := e.cause.Error()
		publicError.Internal = &cause
		if debug {
			publicError.Cause = &cause
		}
	}

	return publicError
}

// defaultFlowError is a struct for flow-related errors.
type defaultFlowError struct {
	defaultError

	status int // The suggested HTTP status code.
}

// createErrorText creates the text used as the string representation of the error.
func createErrorText(code, message string, cause error) string {
	text := fmt.Sprintf("%s - %s", code, message)

	if cause != nil {
		text = fmt.Sprintf("%s: %s", text, cause.Error())
	}

	return text
}

// NewFlowError creates a new FlowError instance.
func NewFlowError(code, message string, status int) FlowError {
	return newFlowErrorWithCause(code, message, status, nil)
}

// newFlowErrorWithCause creates a new FlowError instance with an error cause.
func newFlowErrorWithCause(code, message string, status int, cause error) FlowError {
	errorText := createErrorText(code, message, cause)

	e := defaultError{
		cause:     cause,
		code:      code,
		message:   message,
		errorText: errorText,
	}

	return &defaultFlowError{defaultError: e, status: status}
}

// Status returns the suggested HTTP status code.
func (e *defaultFlowError) Status() int {
	return e.status
}

// Wrap wraps the error with another error.
func (e *defaultFlowError) Wrap(err error) FlowError {
	return newFlowErrorWithCause(e.code, e.message, e.status, err)
}

// defaultInputError is a struct for input-related errors.
type defaultInputError struct {
	defaultError
}

// NewInputError creates a new InputError instance.
func NewInputError(code, message string) InputError {
	return newInputErrorWithCause(code, message, nil)
}

// newInputErrorWithCause creates a new InputError instance with an error cause.
func newInputErrorWithCause(code, message string, cause error) InputError {
	errorText := createErrorText(code, message, cause)

	e := defaultError{
		cause:     cause,
		code:      code,
		message:   message,
		errorText: errorText,
	}

	return &defaultInputError{defaultError: e}
}

// Wrap wraps the error with another error.
func (e *defaultInputError) Wrap(err error) InputError {
	return newInputErrorWithCause(e.code, e.message, err)
}

func createMustBeOneOfError(values []string) InputError {
	return NewInputError("value_invalid_error", fmt.Sprintf("The value is invalid. Must be one of: %s", strings.Join(values, ",")))
}
