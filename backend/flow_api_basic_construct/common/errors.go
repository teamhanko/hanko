package common

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
	"net/http"
)

var (
	ErrorConfigurationError         = flowpilot.NewFlowError("configuration_error", "The configuration contains errors.", http.StatusInternalServerError)
	ErrorDeviceNotCapable           = flowpilot.NewFlowError("device_not_capable", "The device can not login or register.", http.StatusOK) // The device is not able to provide at least one login method.
	ErrorPasscodeInvalid            = flowpilot.NewFlowError("passcode_invalid", "The passcode is invalid.", http.StatusUnauthorized)
	ErrorPasscodeMaxAttemptsReached = flowpilot.NewFlowError("passcode_max_attempts_reached", "The passcode was entered wrong too many times.", http.StatusUnauthorized)
)

var (
	ErrorEmailAlreadyExists    = flowpilot.NewInputError("email_already_exists", "The email address already exists.")
	ErrorUsernameAlreadyExists = flowpilot.NewInputError("username_already_exists", "The username already exists.")
)
