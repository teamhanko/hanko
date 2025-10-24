package flow_locker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/v2/config"
)

func TestNewFlowLocker(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.FlowLocker
		wantType    string // "noop", "memory", "redis", or "error"
		wantErr     bool
		errContains string
	}{
		{
			name: "disabled returns NoOpLocker",
			cfg: config.FlowLocker{
				Enabled: false,
				Store:   config.FLOW_LOCKER_STORE_IN_MEMORY,
			},
			wantType: "noop",
			wantErr:  false,
		},
		{
			name: "in_memory store returns MemoryLocker",
			cfg: config.FlowLocker{
				Enabled: true,
				Store:   config.FLOW_LOCKER_STORE_IN_MEMORY,
			},
			wantType: "memory",
			wantErr:  false,
		},
		{
			name: "redis store with config returns RedisLocker",
			cfg: config.FlowLocker{
				Enabled: true,
				Store:   config.FLOW_LOCKER_STORE_REDIS,
				Redis: &config.RedisConfig{
					Address:  "localhost:6379",
					Password: "secret",
				},
				TTL: 30 * time.Second,
			},
			wantType: "redis",
			wantErr:  false,
		},
		{
			name: "redis store without config returns error",
			cfg: config.FlowLocker{
				Enabled: true,
				Store:   config.FLOW_LOCKER_STORE_REDIS,
				Redis:   nil,
			},
			wantType:    "error",
			wantErr:     true,
			errContains: "redis config required",
		},
		{
			name: "unsupported store type returns error",
			cfg: config.FlowLocker{
				Enabled: true,
				Store:   "unsupported_store",
			},
			wantType:    "error",
			wantErr:     true,
			errContains: "unsupported flow locker store",
		},
		{
			name: "empty store type returns error",
			cfg: config.FlowLocker{
				Enabled: true,
				Store:   "",
			},
			wantType:    "error",
			wantErr:     true,
			errContains: "unsupported flow locker store",
		},
		{
			name: "disabled with invalid config still returns NoOpLocker",
			cfg: config.FlowLocker{
				Enabled: false,
				Store:   "invalid",
			},
			wantType: "noop",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locker, err := NewFlowLocker(tt.cfg)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, locker)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, locker)

			// Check the type of locker returned
			switch tt.wantType {
			case "noop":
				_, ok := locker.(*NoOpLocker)
				assert.True(t, ok, "expected NoOpLocker but got %T", locker)
			case "memory":
				_, ok := locker.(*MemoryLocker)
				assert.True(t, ok, "expected MemoryLocker but got %T", locker)
			case "redis":
				redisLocker, ok := locker.(*RedisLocker)
				assert.True(t, ok, "expected RedisLocker but got %T", locker)

				// Verify configuration was passed correctly
				if ok {
					assert.Equal(t, tt.cfg.TTL, redisLocker.expiry)
				}
			}
		})
	}
}
