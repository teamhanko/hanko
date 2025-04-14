package dto

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"net/http"
	"reflect"
	"strings"
)

type CustomValidator struct {
	Validator *validator.Validate
}

type ValidationErrors struct {
	Errors []string `json:"errors"`
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()

	_ = v.RegisterValidation("hanko_event", webhookEventValidator)

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	return &CustomValidator{Validator: v}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.Validator.Struct(i); err != nil {
		vErrs := TransformValidationErrors(err)
		return echo.NewHTTPError(http.StatusBadRequest, strings.Join(vErrs, " and "))
	}

	return nil
}

func webhookEventValidator(fl validator.FieldLevel) bool {
	return events.StringIsValidEvent(fl.Field().String())
}

func TransformValidationErrors(err error) []string {
	if fieldErrors, ok := err.(validator.ValidationErrors); ok {
		vErrs := make([]string, len(fieldErrors))
		for i, err := range fieldErrors {
			switch err.Tag() {
			case "required":
				vErrs[i] = fmt.Sprintf("%s is a required field", err.Field())
			case "email":
				vErrs[i] = fmt.Sprintf("%s must be a valid email address", err.Field())
			case "uuid":
				vErrs[i] = fmt.Sprintf("%s must be a valid uuid", err.Field())
			case "uuid4":
				vErrs[i] = fmt.Sprintf("%s must be a valid uuid4", err.Field())
			case "url":
				vErrs[i] = fmt.Sprintf("%s must be a valid URL", err.Field())
			case "gte":
				vErrs[i] = fmt.Sprintf("length of %s must be greater or equal to %v", err.Field(), err.Param())
			case "unique":
				vErrs[i] = fmt.Sprintf("%s entries are not unique", err.Field())
			case "hanko_event":
				vErrs[i] = fmt.Sprintf("%s in %s is not a valid webhook event", err.Value(), err.Field())
			case "ip":
				vErrs[i] = fmt.Sprintf("%s must be a valid ip address (v4 or v6)", err.Field())
			case "required_if":
				vErrs[i] = fmt.Sprintf("%s is required if %v", err.Field(), err.Param())
			case "min":
				vErrs[i] = fmt.Sprintf("length of %s must be greater or equal to %v", err.Field(), err.Param())
			case "excluded_if":
				vErrs[i] = fmt.Sprintf("%s must not be set when %s", err.Field(), err.Param())
			default:
				vErrs[i] = fmt.Sprintf("something wrong on %s; %s", err.Field(), err.Tag())
			}
		}
		return vErrs
	}
	return nil
}
