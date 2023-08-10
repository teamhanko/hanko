package flowpilot

import "regexp"

// InputType represents the type of the input field.
type InputType string

// Input types enumeration.
const (
	StringType   InputType = "string"
	EmailType    InputType = "email"
	NumberType   InputType = "number"
	PasswordType InputType = "password"
	JSONType     InputType = "json"
)

// Input defines the interface for input fields.
type Input interface {
	MinLength(minLength int) *DefaultInput
	MaxLength(maxLength int) *DefaultInput
	Required(b bool) *DefaultInput
	Hidden(b bool) *DefaultInput
	Preserve(b bool) *DefaultInput
	Persist(b bool) *DefaultInput

	// TODO: After experimenting with 'ConditionalIncludeFromStash', I realized that it should be replaced with another
	// function 'ConditionalIncludeOnStates(...StateName)'. This function would include the input field when executed
	// under specific states. The issue with 'ConditionalIncludeFromStash' is, that if a method "forgets" to set the
	// value beforehand, the validation won't require the value, since it's not present in the stash. In contrast, using
	// 'ConditionalIncludeOnStates(...)' makes it clear that the value must be present when the current state matches
	// and the decision about whether to validate the value or not isn't dependent on a different method.

	ConditionalIncludeFromStash(b bool) *DefaultInput
	CompareWithStash(b bool) *DefaultInput
	setValue(value interface{}) *DefaultInput
	validate(value *string) bool
}

// defaultExtraInputOptions holds additional input field options.
type defaultExtraInputOptions struct {
	preserveValue               bool
	persistValue                bool
	conditionalIncludeFromStash bool
	compareWithStash            bool
}

// DefaultInput represents an input field with its options.
type DefaultInput struct {
	name      string
	dataType  InputType
	value     interface{}
	minLength *int
	maxLength *int
	required  *bool
	hidden    *bool
	errorType *ErrorType
	defaultExtraInputOptions
}

// PublicInput represents an input field for public exposure.
type PublicInput struct {
	Name      string      `json:"name"`
	Type      InputType   `json:"type"`
	Value     interface{} `json:"value,omitempty"`
	MinLength *int        `json:"min_length,omitempty"`
	MaxLength *int        `json:"max_length,omitempty"`
	Required  *bool       `json:"required,omitempty"`
	Hidden    *bool       `json:"hidden,omitempty"`
	Error     *ErrorType  `json:"error,omitempty"`
}

// newInput creates a new DefaultInput instance with provided parameters.
func newInput(name string, t InputType, persistValue bool) *DefaultInput {
	return &DefaultInput{
		name:     name,
		dataType: t,
		defaultExtraInputOptions: defaultExtraInputOptions{
			preserveValue:               false,
			persistValue:                persistValue,
			conditionalIncludeFromStash: false,
			compareWithStash:            false,
		},
	}
}

// StringInput creates a new input field of string type.
func StringInput(name string) *DefaultInput {
	return newInput(name, StringType, true)
}

// EmailInput creates a new input field of email type.
func EmailInput(name string) *DefaultInput {
	return newInput(name, EmailType, true)
}

// NumberInput creates a new input field of number type.
func NumberInput(name string) *DefaultInput {
	return newInput(name, NumberType, true)
}

// PasswordInput creates a new input field of password type.
func PasswordInput(name string) *DefaultInput {
	return newInput(name, PasswordType, false)
}

// JSONInput creates a new input field of JSON type.
func JSONInput(name string) *DefaultInput {
	return newInput(name, JSONType, false)
}

// MinLength sets the minimum length for the input field.
func (i *DefaultInput) MinLength(minLength int) *DefaultInput {
	i.minLength = &minLength
	return i
}

// MaxLength sets the maximum length for the input field.
func (i *DefaultInput) MaxLength(maxLength int) *DefaultInput {
	i.maxLength = &maxLength
	return i
}

// Required sets whether the input field is required.
func (i *DefaultInput) Required(b bool) *DefaultInput {
	i.required = &b
	return i
}

// Hidden sets whether the input field is hidden.
func (i *DefaultInput) Hidden(b bool) *DefaultInput {
	i.hidden = &b
	return i
}

// Preserve sets whether the input field value should be preserved, so that the value is included in the response
// instead of being blanked out.
func (i *DefaultInput) Preserve(b bool) *DefaultInput {
	i.preserveValue = b
	return i
}

// Persist sets whether the input field value should be persisted.
func (i *DefaultInput) Persist(b bool) *DefaultInput {
	i.persistValue = b
	return i
}

// ConditionalIncludeFromStash sets whether the input field is conditionally included from the stash.
func (i *DefaultInput) ConditionalIncludeFromStash(b bool) *DefaultInput {
	i.conditionalIncludeFromStash = b
	return i
}

// CompareWithStash sets whether the input field is compared with stash values.
func (i *DefaultInput) CompareWithStash(b bool) *DefaultInput {
	i.compareWithStash = b
	return i
}

// setValue sets the value for the input field.
func (i *DefaultInput) setValue(value interface{}) *DefaultInput {
	i.value = &value
	return i
}

// validate performs validation on the input field.
func (i *DefaultInput) validate(inputValue *string, stashValue *string) bool {
	// Validate based on input field options.

	// TODO: Replace with more structured validation logic.

	if i.conditionalIncludeFromStash && stashValue == nil {
		return true
	}

	if i.required != nil && *i.required && (inputValue == nil || len(*inputValue) <= 0) {
		i.errorType = ValueMissingError
		return false
	}

	if i.compareWithStash && inputValue != nil && stashValue != nil && *inputValue != *stashValue {
		i.errorType = ValueInvalidError
		return false
	}

	if i.dataType == JSONType {
		// skip further validation
		return true
	}

	if i.minLength != nil {
		if len(*inputValue) < *i.minLength {
			i.errorType = ValueTooShortError
			return false
		}
	}

	if i.maxLength != nil {
		if len(*inputValue) > *i.maxLength {
			i.errorType = ValueTooLongError
			return false
		}
	}

	if i.dataType == EmailType {
		pattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if matched := pattern.MatchString(*inputValue); !matched {
			i.errorType = EmailInvalidError
			return false
		}
	}

	return true
}

// toPublicInput converts the DefaultInput to a PublicInput for public exposure.
func (i *DefaultInput) toPublicInput() *PublicInput {
	return &PublicInput{
		Name:      i.name,
		Type:      i.dataType,
		Value:     i.value,
		MinLength: i.minLength,
		MaxLength: i.maxLength,
		Required:  i.required,
		Hidden:    i.hidden,
		Error:     i.errorType,
	}
}
