package flow_locker

import (
	"context"

	"github.com/gofrs/uuid"
)

// NoOpLocker is a no-op implementation that doesn't actually lock
type NoOpLocker struct{}

// NewNoOpLocker creates a new no-op locker
func NewNoOpLocker() *NoOpLocker {
	return &NoOpLocker{}
}

// Lock does nothing and returns a no-op unlock function
func (n *NoOpLocker) Lock(ctx context.Context, flowID uuid.UUID) (func(), error) {
	return func() {}, nil
}
