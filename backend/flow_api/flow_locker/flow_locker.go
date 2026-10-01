package flow_locker

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/config"
)

// FlowLocker provides an interface for locking flow execution by flow ID
type FlowLocker interface {
	// Lock acquires a lock for the given flow ID
	// Returns an unlock function that must be called when done
	Lock(ctx context.Context, flowID uuid.UUID) (unlock func(context.Context) error, err error)
}

// NewFlowLocker creates a FlowLocker based on configuration
func NewFlowLocker(cfg config.FlowLocker) (FlowLocker, error) {
	if !cfg.Enabled {
		return NewNoOpLocker(), nil
	}

	switch cfg.Store {
	case config.FLOW_LOCKER_STORE_REDIS:
		address, database, err := cfg.Redis.DialAddress()
		if err != nil {
			return nil, err
		}
		return NewRedisLocker(RedisLockerConfig{
			Address:  address,
			Password: cfg.Redis.Password,
			Database: database,
			Expiry:   cfg.TTL,
		}), nil
	case config.FLOW_LOCKER_STORE_IN_MEMORY:
		return NewMemoryLocker(), nil
	default:
		return nil, fmt.Errorf("unsupported flow locker store: %s", cfg.Store)
	}
}
