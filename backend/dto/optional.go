package dto

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// OptionalString represents a PATCH-able string field with 3 states:
// - not present in JSON => Present=false (no change)
// - present with string => Present=true, Value!=nil (set)
// - present with null   => Present=true, Value==nil (clear)
type OptionalString struct {
	Present bool
	Value   *string
}

func (o *OptionalString) UnmarshalJSON(b []byte) error {
	o.Present = true

	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		o.Value = nil
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("expected string or null: %w", err)
	}

	o.Value = &s
	return nil
}
