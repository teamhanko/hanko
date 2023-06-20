package dto

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
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
		if fieldErrors, ok := err.(validator.ValidationErrors); ok {
			vErrs := make([]string, len(fieldErrors))
			for i, err := range fieldErrors {
				switch err.Tag() {
				case "required":
					vErrs[i] = fmt.Sprintf("%s is a required field", err.Field())
				case "email":
					vErrs[i] = fmt.Sprintf("%s must be a valid email address", err.Field())
				case "uuid4":
					vErrs[i] = fmt.Sprintf("%s must be a valid uuid4", err.Field())
				case "url":
					vErrs[i] = fmt.Sprintf("%s must be a valid URL", err.Field())
				default:
					vErrs[i] = fmt.Sprintf("something wrong on %s; %s", err.Field(), err.Tag())
				}
			}

			return echo.NewHTTPError(http.StatusBadRequest, strings.Join(vErrs, " and "))
		}
	}

	return nil
}
