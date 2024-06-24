package flowpilot

type allowedValue interface {
	toPublicAllowedValue() *ResponseAllowedValue
	getValue() interface{}
}

type defaultAllowedValue struct {
	value interface{}
	text  string
}

func (av *defaultAllowedValue) getValue() interface{} {
	return av.value
}

// toPublicAllowedValue converts the allowedValue to a ResponseAllowedValue for public exposure.
func (av *defaultAllowedValue) toPublicAllowedValue() *ResponseAllowedValue {
	return &ResponseAllowedValue{
		Value: av.value,
		Text:  av.text,
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
	l := make([]string, len(*av))
	for i, v := range *av {
		l[i] = v.getValue().(string)
	}
	return l
}

func (av *defaultAllowedValues) toPublicAllowedValues() *ResponseAllowedValues {
	values := make(ResponseAllowedValues, len(*av))
	for i, v := range *av {
		values[i] = v.toPublicAllowedValue()
	}
	return &values
}
