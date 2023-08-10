package utils

import (
	"fmt"
	"github.com/gofrs/uuid"
	"strings"
)

// ParsedAction represents a parsed action from an input string.
type ParsedAction struct {
	MethodName string    // The name of the method extracted from the input string.
	FlowID     uuid.UUID // The UUID representing the flow ID extracted from the input string.
}

// ParseActionParam parses an input string to extract method name and flow ID.
func ParseActionParam(inputString string) (*ParsedAction, error) {
	if inputString == "" {
		return nil, fmt.Errorf("input string is empty")
	}

	// Split the input string into method and flow ID parts using "@" as separator.
	parts := strings.SplitN(inputString, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid input string format")
	}

	// Extract method name from the first part of the split.
	method := parts[0]
	if len(method) == 0 {
		return nil, fmt.Errorf("first part of input string is empty")
	}

	// Parse the second part of the input string into a UUID representing the flow ID.
	flowID, err := uuid.FromString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse second part of the input string: %w", err)
	}

	// Return a ParsedAction instance with extracted method name and flow ID.
	return &ParsedAction{MethodName: method, FlowID: flowID}, nil
}

// CreateActionParam creates an input string from method name and flow ID.
func CreateActionParam(method string, flowID uuid.UUID) string {
	return fmt.Sprintf("%s@%s", method, flowID)
}
