package flow_locker

import (
	"context"
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
)

// MemoryLocker implements FlowLocker using fail-fast locking (i.e. concurrent requests do not wait)
type MemoryLocker struct {
	mu    sync.Mutex
	locks map[uuid.UUID]bool
}

// NewMemoryLocker creates a new in-memory flow locker
func NewMemoryLocker() *MemoryLocker {
	return &MemoryLocker{
		locks: make(map[uuid.UUID]bool),
	}
}

// Lock tries to acquire a lock for the given flow ID
// Returns error immediately if lock is already held.
func (m *MemoryLocker) Lock(ctx context.Context, flowID uuid.UUID) (func(context.Context) error, error) {
	// Check if context is already canceled before attempting lock
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.locks[flowID] {
		return nil, fmt.Errorf("flow %s is already being processed", flowID)
	}

	m.locks[flowID] = true

	unlock := func(ctx context.Context) error {
		m.mu.Lock()
		delete(m.locks, flowID)
		m.mu.Unlock()

		return nil
	}

	return unlock, nil
}
