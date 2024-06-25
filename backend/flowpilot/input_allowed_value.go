package flowpilot

type allowedValue interface {
	toPublicAllowedValue() *ResponseAllowedValue
	getValue() interface{}
}

type defaultAllowedValue struct {
	text  string
	value interface{}
}

func (av *defaultAllowedValue) getValue() interface{} {
	return av.value
}

// toPublicAllowedValue converts the allowedValue to a ResponseAllowedValue for public exposure.
func (av *defaultAllowedValue) toPublicAllowedValue() *ResponseAllowedValue {
	return &ResponseAllowedValue{
		Text:  av.text,
		Value: av.value,
	}
}

type allowedValues interface {
	isAllowed(value string) bool
	add(allowedValue)
	toPublicAllowedValues() *ResponseAllowedValues
	hasAny() bool
	getValues() []string
}

type defaultAllowedValues []allowedValue

func (av *defaultAllowedValues) isAllowed(value string) bool {
	if len(*av) == 0 {
		return true
	}

	for _, v := range *av {
		if v.getValue().(string) == value {
			return true
		}
	}

	return false
}

func (av *defaultAllowedValues) add(value allowedValue) {
	*av = append(*av, value)
}

func (av *defaultAllowedValues) hasAny() bool {
	return len(*av) > 0
}

func (av *defaultAllowedValues) getValues() []string {
	values := make([]string, len(*av))
	for i, v := range *av {
		values[i] = v.getValue().(string)
	}
	return values
}

func (av *defaultAllowedValues) toPublicAllowedValues() *ResponseAllowedValues {
	values := make(ResponseAllowedValues, len(*av))
	for i, v := range *av {
		values[i] = v.toPublicAllowedValue()
	}
	return &values
}
