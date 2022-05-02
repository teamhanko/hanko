package dto

import "net/http"

type ApiError struct {
	Code             int      `json:"code"`
	Message          string   `json:"message"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
}

func NewApiError(code int) *ApiError {
	return &ApiError{
		Code:    code,
		Message: http.StatusText(code),
	}
}

func (e *ApiError) WithValidationErrors(validationErrors []string) *ApiError {
	e.ValidationErrors = validationErrors
	return e
}

func (e *ApiError) WithMessage(message string) *ApiError {
	e.Message = message
	return e
}

func (e *ApiError) Error() string {
	return e.Message
}
