package flowpilot

import "net/http"

// Predefined flow error types
var (
	ErrorTechnical             = NewFlowError("technical_error", "Something went wrong.", http.StatusInternalServerError)
	ErrorFlowExpired           = NewFlowError("flow_expired_error", "The flow has expired.", http.StatusGone)
	ErrorFlowDiscontinuity     = NewFlowError("flow_discontinuity_error", "The flow can't be continued.", http.StatusInternalServerError)
	ErrorOperationNotPermitted = NewFlowError("operation_not_permitted_error", "The operation is not permitted.", http.StatusForbidden)
	ErrorFormDataInvalid       = NewFlowError("form_data_invalid_error", "Form data invalid.", http.StatusBadRequest)
)

// Predefined input error types
var (
	ErrorValueMissing  = NewInputError("value_missing_error", "The value is missing.")
	ErrorValueInvalid  = NewInputError("value_invalid_error", "The value is invalid.")
	ErrorValueTooLong  = NewInputError("value_too_long_error", "The value is too long.")
	ErrorValueTooShort = NewInputError("value_too_short_error", "The value is too short.")
)
