package flowpilot

import (
	"fmt"
	"github.com/gofrs/uuid"
	"strings"
)

// parsedQueryParam represents a parsed action from an input string.
type parsedQueryParam struct {
	actionName ActionName // The name of the action extracted from the input string.
	flowID     uuid.UUID  // The UUID representing the flow ID extracted from the input string.
}

// parseQueryParam parses an input string to extract action name and flow ID.
func parseQueryParam(inputString string) (*parsedQueryParam, error) {
	if inputString == "" {
		return nil, fmt.Errorf("input string is empty")
	}

	// Split the input string into action and flow ID parts using "@" as separator.
	parts := strings.SplitN(inputString, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid input string format")
	}

	// Extract action name from the first part of the split.
	action := parts[0]
	if len(action) == 0 {
		return nil, fmt.Errorf("first part of input string is empty")
	}

	// Parse the second part of the input string into a UUID representing the flow ID.
	flowID, err := uuid.FromString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse second part of the input string: %w", err)
	}

	// Return a parsedQueryParam instance with extracted action name and flow ID.
	return &parsedQueryParam{actionName: ActionName(action), flowID: flowID}, nil
}

// createQueryParam creates an input string from action name and flow ID.
func createQueryParam(action string, flowID uuid.UUID) string {
	return fmt.Sprintf("%s@%s", action, flowID)
}
