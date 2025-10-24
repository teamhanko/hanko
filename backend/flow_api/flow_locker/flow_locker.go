package flow_locker

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/config"
)

// FlowLocker provides an interface for locking flow execution by flow ID
type FlowLocker interface {
	// Lock acquires a lock for the given flow ID
	// Returns an unlock function that must be called when done
	Lock(ctx context.Context, flowID uuid.UUID) (unlock func(), err error)
}

// NewFlowLocker creates a FlowLocker based on configuration
func NewFlowLocker(cfg config.FlowLocker) (FlowLocker, error) {
	if !cfg.Enabled {
		return NewNoOpLocker(), nil
	}

	switch cfg.Store {
	case config.FLOW_LOCKER_STORE_REDIS:
		if cfg.Redis == nil {
			return nil, fmt.Errorf("redis config required for redis flow locker")
		}
		return NewRedisLocker(RedisLockerConfig{
			Address:  cfg.Redis.Address,
			Password: cfg.Redis.Password,
			Expiry:   cfg.TTL,
		}), nil
	case config.FLOW_LOCKER_STORE_IN_MEMORY:
		return NewMemoryLocker(), nil
	default:
		return nil, fmt.Errorf("unsupported flow locker store: %s", cfg.Store)
	}
}
