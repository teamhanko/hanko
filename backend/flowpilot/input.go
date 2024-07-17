package flowpilot

import (
	"regexp"
)

// inputType represents the type of the input field.
type inputType string

// Input types enumeration.
const (
	inputTypeString   inputType = "string"
	inputTypeBoolean  inputType = "boolean"
	inputTypeEmail    inputType = "email"
	inputTypeNumber   inputType = "number"
	inputTypePassword inputType = "password"
	inputTypeJSON     inputType = "json"
)

// Input defines the interface for input fields.
type Input interface {
	MinLength(minLength int) Input
	MaxLength(maxLength int) Input
	Required(b bool) Input
	Hidden(b bool) Input
	Preserve(b bool) Input
	AllowedValue(name string, value interface{}) Input
	TrimSpace(b bool) Input
	LowerCase(b bool) Input

	setValue(value interface{}) Input
	setError(inputError InputError)
	getError() InputError
	getName() string
	shouldPreserve() bool
	shouldTrimSpace() bool
	shouldConvertToLowerCase() bool
	validate(inputData readOnlyActionInput) bool
	toResponseInput() *ResponseInput
}

// defaultExtraInputOptions holds additional input field options.
type defaultExtraInputOptions struct {
	preserveValue bool
	trimSpace     bool
	lowerCase     bool
}

// defaultInput represents an input field with its options.
type defaultInput struct {
	name          string
	dataType      inputType
	value         interface{}
	minLength     *int
	maxLength     *int
	required      *bool
	hidden        *bool
	error         InputError
	allowedValues allowedValues

	defaultExtraInputOptions
}

func (i *defaultInput) AllowedValue(name string, value interface{}) Input {
	i.allowedValues.add(&defaultAllowedValue{
		value: value,
		text:  name,
	})
	return i
}

// newInput creates a new input instance with provided parameters.
func newInput(name string, inputType inputType) Input {
	extraOptions := defaultExtraInputOptions{
		preserveValue: false,
		trimSpace:     false,
		lowerCase:     false,
	}

	return &defaultInput{
		name:                     name,
		dataType:                 inputType,
		defaultExtraInputOptions: extraOptions,
		allowedValues:            &defaultAllowedValues{},
	}
}

// StringInput creates a new input field of string type.
func StringInput(name string) Input {
	return newInput(name, inputTypeString)
}

// EmailInput creates a new input field of email type.
func EmailInput(name string) Input {
	return newInput(name, inputTypeEmail)
}

// NumberInput creates a new input field of number type.
func NumberInput(name string) Input {
	return newInput(name, inputTypeNumber)
}

// BooleanInput creates a new input field of boolean type.
func BooleanInput(name string) Input {
	return newInput(name, inputTypeBoolean)
}

// PasswordInput creates a new input field of password type.
func PasswordInput(name string) Input {
	return newInput(name, inputTypePassword)
}

// JSONInput creates a new input field of JSON type.
func JSONInput(name string) Input {
	return newInput(name, inputTypeJSON)
}

// MinLength sets the minimum length for the input field.
func (i *defaultInput) MinLength(minLength int) Input {
	i.minLength = &minLength
	return i
}

// MaxLength sets the maximum length for the input field.
func (i *defaultInput) MaxLength(maxLength int) Input {
	i.maxLength = &maxLength
	return i
}

// Required sets whether the input field is required.
func (i *defaultInput) Required(b bool) Input {
	i.required = &b
	return i
}

// Hidden sets whether the input field is hidden.
func (i *defaultInput) Hidden(b bool) Input {
	i.hidden = &b
	return i
}

// Preserve sets whether the input field value should be preserved, so that the value is included in the response
// instead of being blanked out.
func (i *defaultInput) Preserve(b bool) Input {
	i.preserveValue = b
	return i
}

// TrimSpace sets whether the leading and trailing whitespaces should be trimmed.
func (i *defaultInput) TrimSpace(b bool) Input {
	i.trimSpace = b
	return i
}

// LowerCase sets whether the value should be converted to lower case.
func (i *defaultInput) LowerCase(b bool) Input {
	i.lowerCase = b
	return i
}

// setValue sets the value for the input field for the current response.
func (i *defaultInput) setValue(value interface{}) Input {
	i.value = &value
	return i
}

// getName returns the name of the input field.
func (i *defaultInput) getName() string {
	return i.name
}

// setError sets an error to the given input field.
func (i *defaultInput) setError(inputError InputError) {
	i.error = inputError
}

// getError returns the input error.
func (i *defaultInput) getError() InputError {
	return i.error
}

// shouldPersist indicates the value should be preserved.
func (i *defaultInput) shouldPreserve() bool {
	return i.preserveValue
}

// shouldPersist indicates the value should be preserved.
func (i *defaultInput) shouldTrimSpace() bool {
	return i.trimSpace
}

// shouldConvertToLowerCase indicates the value should be converted to lower case.
func (i *defaultInput) shouldConvertToLowerCase() bool {
	return i.lowerCase
}

// validate performs validation on the input field.
func (i *defaultInput) validate(inputData readOnlyActionInput) bool {
	// TODO: Replace with more structured validation logic.

	var inputValue *string

	if v := inputData.Get(i.name); v.Exists() {
		inputValue = &v.Str
	}

	if i.dataType == inputTypeJSON {
		// skip further validation
		return true
	}

	if i.dataType == inputTypeBoolean {
		return true
	}

	isRequired := i.required != nil && *i.required
	hasEmptyOrNilValue := inputValue == nil || len(*inputValue) <= 0

	if isRequired && hasEmptyOrNilValue {
		i.error = ErrorValueMissing
		return false
	}

	if !hasEmptyOrNilValue && i.minLength != nil {
		if len(*inputValue) < *i.minLength {
			i.error = ErrorValueTooShort
			return false
		}
	}

	if !hasEmptyOrNilValue && i.maxLength != nil {
		if len(*inputValue) > *i.maxLength {
			i.error = ErrorValueTooLong
			return false
		}
	}

	if i.dataType == inputTypeEmail && (isRequired || (!isRequired && !hasEmptyOrNilValue)) {
		pattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if matched := pattern.MatchString(*inputValue); !matched {
			i.error = ErrorValueInvalid
			return false
		}
	}

	if i.dataType == inputTypeString && inputValue != nil {
		if i.allowedValues.isAllowed(*inputValue) {
			return true
		}

		i.error = createMustBeOneOfError(i.allowedValues.getValues())
		return false
	}

	return true
}

// toResponseInput converts the defaultInput to a ResponseInput for public exposure.
func (i *defaultInput) toResponseInput() *ResponseInput {
	var e *ResponseError
	var av *ResponseAllowedValues

	if i.error != nil {
		e = i.error.toResponseError(true)
	}

	if i.allowedValues != nil && i.allowedValues.hasAny() {
		av = i.allowedValues.toResponseAllowedValues()
	}

	return &ResponseInput{
		Name:          i.name,
		Type:          i.dataType,
		Value:         i.value,
		MinLength:     i.minLength,
		MaxLength:     i.maxLength,
		Required:      i.required,
		Hidden:        i.hidden,
		Error:         e,
		AllowedValues: av,
	}
}
