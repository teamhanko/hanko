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
	Persist(b bool) Input
	ConditionalIncludeOnState(stateNames ...StateName) Input
	CompareWithStash(b bool) Input
	AllowedValue(value interface{}, name string) Input

	setValue(value interface{}) Input
	setError(inputError InputError)
	getName() string
	shouldPersist() bool
	shouldPreserve() bool
	isIncludedOnState(stateName StateName) bool
	validate(stateName StateName, inputData readOnlyActionInput, stashData stash) bool
	toPublicInput() *PublicInput
}

// defaultExtraInputOptions holds additional input field options.
type defaultExtraInputOptions struct {
	preserveValue    bool
	persistValue     bool
	includeOnStates  []StateName
	compareWithStash bool
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

func (i *defaultInput) AllowedValue(value interface{}, name string) Input {
	i.allowedValues = append(i.allowedValues, &allowedValue{
		value: value,
		text:  name,
	})
	return i
}

// newInput creates a new input instance with provided parameters.
func newInput(name string, inputType inputType, persistValue bool) Input {
	extraOptions := defaultExtraInputOptions{
		preserveValue:    false,
		persistValue:     persistValue,
		includeOnStates:  []StateName{},
		compareWithStash: false,
	}

	return &defaultInput{
		name:                     name,
		dataType:                 inputType,
		defaultExtraInputOptions: extraOptions,
	}
}

type allowedValue struct {
	value interface{}
	text  string
}

// toPublicAllowedValue converts the allowedValue to a PublicAllowedValue for public exposure.
func (av *allowedValue) toPublicAllowedValue() *PublicAllowedValue {
	return &PublicAllowedValue{
		Value: av.value,
		Text:  av.text,
	}
}

type allowedValues []*allowedValue

func (av *allowedValues) values() []interface{} {
	var values []interface{}
	for _, v := range *av {
		values = append(values, v.value)
	}
	return values
}

func (av *allowedValues) toPublicAllowedValues() *PublicAllowedValues {
	var values PublicAllowedValues
	for _, v := range *av {
		values = append(values, v.toPublicAllowedValue())
	}
	return &values
}

// StringInput creates a new input field of string type.
func StringInput(name string) Input {
	return newInput(name, inputTypeString, true)
}

// EmailInput creates a new input field of email type.
func EmailInput(name string) Input {
	return newInput(name, inputTypeEmail, true)
}

// NumberInput creates a new input field of number type.
func NumberInput(name string) Input {
	return newInput(name, inputTypeNumber, true)
}

// BooleanInput creates a new input field of boolean type.
func BooleanInput(name string) Input {
	return newInput(name, inputTypeBoolean, true)
}

// PasswordInput creates a new input field of password type.
func PasswordInput(name string) Input {
	return newInput(name, inputTypePassword, false)
}

// JSONInput creates a new input field of JSON type.
func JSONInput(name string) Input {
	return newInput(name, inputTypeJSON, false)
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

// Persist sets whether the input field value should be persisted.
func (i *defaultInput) Persist(b bool) Input {
	i.persistValue = b
	return i
}

// ConditionalIncludeOnState sets the states where the input field is included.
func (i *defaultInput) ConditionalIncludeOnState(stateNames ...StateName) Input {
	i.includeOnStates = stateNames
	return i
}

// isIncludedOnState check if a conditional input field is included according to the given stateName.
func (i *defaultInput) isIncludedOnState(stateName StateName) bool {
	if len(i.includeOnStates) == 0 {
		return true
	}

	for _, s := range i.includeOnStates {
		if s == stateName {
			return true
		}
	}

	return false
}

// CompareWithStash sets whether the input field is compared with stash values.
func (i *defaultInput) CompareWithStash(b bool) Input {
	i.compareWithStash = b
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

// shouldPersist indicates the value should be persisted.
func (i *defaultInput) shouldPersist() bool {
	return i.persistValue
}

// shouldPersist indicates the value should be preserved.
func (i *defaultInput) shouldPreserve() bool {
	return i.preserveValue
}

// validate performs validation on the input field.
func (i *defaultInput) validate(stateName StateName, inputData readOnlyActionInput, stashData stash) bool {
	// TODO: Replace with more structured validation logic.

	var inputValue *string
	var stashValue *string

	if v := inputData.Get(i.name); v.Exists() {
		inputValue = &v.Str
	}

	if v := stashData.Get(i.name); v.Exists() {
		stashValue = &v.Str
	}

	if len(i.includeOnStates) > 0 && !i.isIncludedOnState(stateName) {
		// skip validation
		return true
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

	if i.compareWithStash && inputValue != nil && stashValue != nil && *inputValue != *stashValue {
		i.error = ErrorValueInvalid
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
			i.error = ErrorEmailInvalid
			return false
		}
	}

	if i.dataType == inputTypeString && i.allowedValues != nil {
		for _, v := range i.allowedValues.values() {
			if v.(string) == *inputValue {
				return true
			}
		}
		i.error = createMustBeOneOfError(i.allowedValues.values())
		return false
	}

	return true
}

// toPublicInput converts the defaultInput to a PublicInput for public exposure.
func (i *defaultInput) toPublicInput() *PublicInput {
	var publicError *PublicError

	if i.error != nil {
		e := i.error.toPublicError(true)
		publicError = &e
	}

	return &PublicInput{
		Name:          i.name,
		Type:          i.dataType,
		Value:         i.value,
		MinLength:     i.minLength,
		MaxLength:     i.maxLength,
		Required:      i.required,
		Hidden:        i.hidden,
		PublicError:   publicError,
		AllowedValues: i.allowedValues.toPublicAllowedValues(),
	}
}
