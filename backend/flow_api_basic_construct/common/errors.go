package common

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
	"net/http"
)

var ErrorDeviceNotCapable = flowpilot.NewFlowError("device_not_capable", "The device can not login or register.", http.StatusOK) // The device is not able to provide at least one login method.

var (
	ErrorEmailAlreadyExists    = flowpilot.NewInputError("email_already_exists", "The email address already exists.")
	ErrorUsernameAlreadyExists = flowpilot.NewInputError("username_already_exists", "The username already exists.")
)
