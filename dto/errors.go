package dto

import "net/http"

type ApiError struct {
	Code    int
	Message string
}

func NewApiError(code int) *ApiError {
	return &ApiError{
		Code:    code,
		Message: http.StatusText(code),
	}
}
