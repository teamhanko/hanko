package thirdparty

import (
	"fmt"
	"net/url"
)

type ThirdPartyError struct {
	Code        string
	Description string
	Cause       error
}

func (e *ThirdPartyError) Query() string {
	q := url.Values{}
	q.Add("error", e.Code)

	if e.Code == ErrorCodeServerError {
		q.Add("error_description", "an internal error has occurred")
	} else if e.Description != "" {
		q.Add("error_description", e.Description)
	}
	return q.Encode()
}

func (e *ThirdPartyError) WithDescription(description string) *ThirdPartyError {
	e.Description = description
	return e
}

func (e *ThirdPartyError) WithCause(cause error) *ThirdPartyError {
	e.Cause = cause
	return e
}

func (e *ThirdPartyError) Error() string {
	err := fmt.Sprintf("thirdparty: %s", e.Code)

	if e.Description != "" {
		err = fmt.Sprintf("%s: %s", err, e.Description)
	}

	if e.Cause != nil {
		return fmt.Sprintf("%s: %s", err, e.Cause)
	}
	return err
}

func NewThirdPartyError(code string, description string) *ThirdPartyError {
	return &ThirdPartyError{Code: code, Description: description}
}

func ErrorInvalidRequest(desc string) *ThirdPartyError {
	return &ThirdPartyError{Code: ErrorCodeInvalidRequest, Description: desc}
}

func ErrorServer(desc string) *ThirdPartyError {
	return &ThirdPartyError{Code: ErrorCodeServerError, Description: desc}
}

func ErrorUserConflict(desc string) *ThirdPartyError {
	return &ThirdPartyError{Code: ErrorCodeUserConflict, Description: desc}
}

func ErrorMultipleAccounts(desc string) *ThirdPartyError {
	return &ThirdPartyError{Code: ErrorCodeMultipleAccounts, Description: desc}
}

func ErrorUnverifiedProviderEmail(desc string) *ThirdPartyError {
	return &ThirdPartyError{Code: ErrorCodeUnverifiedProviderEmail, Description: desc}
}

func ErrorMaxNumberOfAddresses(desc string) *ThirdPartyError {
	return &ThirdPartyError{Code: ErrorCodeMaxNumberOfAddresses, Description: desc}
}

const (
	ErrorCodeInvalidRequest          = "invalid_request"
	ErrorCodeServerError             = "server_error"
	ErrorCodeUserConflict            = "user_conflict"
	ErrorCodeMultipleAccounts        = "multiple_accounts"
	ErrorCodeUnverifiedProviderEmail = "unverified_email"
	ErrorCodeMaxNumberOfAddresses    = "email_maxnum"
)
