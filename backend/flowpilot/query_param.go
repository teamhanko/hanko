package flowpilot

import (
	"fmt"
	"github.com/gofrs/uuid"
	"net/url"
	"strings"
)

type queryParam interface {
	getKey() string
	getValue() string
	getActionName() ActionName
	getFlowID() uuid.UUID
	getURLValues() url.Values
}

// parsedQueryParamValue represents a parsed action from an input string.
type parsedQueryParamValue struct {
	actionName ActionName // The actionName of the action extracted from the input string.
	flowID     uuid.UUID  // The UUID representing the flow ID extracted from the input string.
}

// defaultQueryParam represents a parsed action from an input string.
type defaultQueryParam struct {
	key string

	*parsedQueryParamValue
}

func createQueryParamValue(actionName ActionName, flowID uuid.UUID) string {
	return fmt.Sprintf("%s@%s", actionName, flowID)
}

// parseValue parses an input string to extract action name and flow ID.
func parseQueryParamValue(value string) (*parsedQueryParamValue, error) {
	if value == "" {
		return nil, fmt.Errorf("query param value is empty")
	}

	// Split the input string into action and flow ID parts using "@" as separator.
	parts := strings.SplitN(value, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid query param value format")
	}

	// Extract action name from the first part of the split.
	action := parts[0]
	if len(action) == 0 {
		return nil, fmt.Errorf("first part of the query param value is empty")
	}

	// Parse the second part of the input string into a UUID representing the flow ID.
	flowID, err := uuid.FromString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse second part of the query param value: %w", err)
	}

	// Return a defaultQueryParam instance with extracted action name and flow ID.
	return &parsedQueryParamValue{actionName: ActionName(action), flowID: flowID}, nil
}

func newQueryParam(key, value string) (queryParam, error) {
	v, err := parseQueryParamValue(value)
	return &defaultQueryParam{key: key, parsedQueryParamValue: v}, err
}

func (q *defaultQueryParam) getKey() string {
	return q.key
}

func (q *defaultQueryParam) getValue() string {
	return createQueryParamValue(q.getActionName(), q.getFlowID())
}

func (q *defaultQueryParam) getURLValues() url.Values {
	values := url.Values{}
	values.Add(q.getKey(), q.getValue())
	return values
}

func (q *defaultQueryParam) getActionName() ActionName {
	return q.parsedQueryParamValue.actionName
}

func (q *defaultQueryParam) getFlowID() uuid.UUID {
	return q.parsedQueryParamValue.flowID
}
