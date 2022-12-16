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

	if e.Code == ThirdPartyErrorCodeServerError {
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
	return &ThirdPartyError{Code: ThirdPartyErrorCodeInvalidRequest, Description: desc}
}

func ErrorServer(desc string) *ThirdPartyError {
	return &ThirdPartyError{Code: ThirdPartyErrorCodeServerError, Description: desc}
}

func ErrorUserConflict(desc string) *ThirdPartyError {
	return &ThirdPartyError{Code: ThirdPartyErrorCodeUserConflict, Description: desc}
}

func ErrorMultipleAccounts(desc string) *ThirdPartyError {
	return &ThirdPartyError{Code: ThirdPartyErrorCodeMultipleAccounts, Description: desc}
}

func ErrorUnverifiedProviderEmail(desc string) *ThirdPartyError {
	return &ThirdPartyError{Code: ThirdPartyErrorUnverifiedProviderEmail, Description: desc}
}

func ErrorMaxNumberOfAddresses(desc string) *ThirdPartyError {
	return &ThirdPartyError{Code: ThirdPartyErrorMaxNumberOfAddresses, Description: desc}
}

const (
	ThirdPartyErrorCodeInvalidRequest      = "invalid_request"
	ThirdPartyErrorCodeServerError         = "server_error"
	ThirdPartyErrorCodeUserConflict        = "user_conflict"
	ThirdPartyErrorCodeMultipleAccounts    = "multiple_accounts"
	ThirdPartyErrorUnverifiedProviderEmail = "unverified_email"
	ThirdPartyErrorMaxNumberOfAddresses    = "email_maxnum"
)
