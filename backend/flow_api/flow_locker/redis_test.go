package flow_locker

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RedisLockerTestSuite struct {
	suite.Suite
	pool         *dockertest.Pool
	resource     *dockertest.Resource
	redisAddress string
}

func (suite *RedisLockerTestSuite) SetupSuite() {
	var err error

	suite.pool, err = dockertest.NewPool("")
	require.NoError(suite.T(), err, "Could not construct pool")

	err = suite.pool.Client.Ping()
	require.NoError(suite.T(), err, "Could not connect to Docker")

	suite.resource, err = suite.pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "redis",
		Tag:        "8-alpine",
		Env:        []string{},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(suite.T(), err, "Could not start Redis container")

	_ = suite.resource.Expire(600)

	// Wait for Redis to be ready
	suite.redisAddress = fmt.Sprintf("localhost:%s", suite.resource.GetPort("6379/tcp"))

	suite.pool.MaxWait = 30 * time.Second
	err = suite.pool.Retry(func() error {
		locker := NewRedisLocker(RedisLockerConfig{
			Address: suite.redisAddress,
			Expiry:  10 * time.Second,
		})

		ctx := context.Background()
		testID := uuid.Must(uuid.NewV4())
		unlock, err := locker.Lock(ctx, testID)
		if err != nil {
			return err
		}
		return unlock(ctx)
	})
	require.NoError(suite.T(), err, "Could not connect to Redis")
}

func (suite *RedisLockerTestSuite) TearDownSuite() {
	if suite.resource != nil {
		err := suite.pool.Purge(suite.resource)
		require.NoError(suite.T(), err, "Could not purge Redis container")
	}
}

func (suite *RedisLockerTestSuite) getTestRedisLocker() *RedisLocker {
	return NewRedisLocker(RedisLockerConfig{
		Address: suite.redisAddress,
		Expiry:  10 * time.Second,
	})
}

// TestRedisLockerSuite runs the test suite
func TestRedisLockerSuite(t *testing.T) {
	suite.Run(t, new(RedisLockerTestSuite))
}

func (suite *RedisLockerTestSuite) TestLock_Success() {
	locker := suite.getTestRedisLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	unlock, err := locker.Lock(ctx, flowID)

	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), unlock)

	err = unlock(ctx)
	assert.NoError(suite.T(), err, "unlock should succeed")
}

func (suite *RedisLockerTestSuite) TestLock_FailFast_WhenAlreadyLocked() {
	locker := suite.getTestRedisLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	unlock1, err := locker.Lock(ctx, flowID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), unlock1)
	defer func() {
		err := unlock1(ctx)
		assert.NoError(suite.T(), err)
	}()

	unlock2, err := locker.Lock(ctx, flowID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), unlock2)
}

func (suite *RedisLockerTestSuite) TestLock_SucceedsAfterUnlock() {
	locker := suite.getTestRedisLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	unlock1, err := locker.Lock(ctx, flowID)
	require.NoError(suite.T(), err)
	err = unlock1(ctx)
	require.NoError(suite.T(), err)

	time.Sleep(10 * time.Millisecond)

	unlock2, err := locker.Lock(ctx, flowID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), unlock2)
	err = unlock2(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *RedisLockerTestSuite) TestLock_DifferentFlowIDs() {
	locker := suite.getTestRedisLocker()
	ctx := context.Background()
	flowID1 := uuid.Must(uuid.NewV4())
	flowID2 := uuid.Must(uuid.NewV4())

	unlock1, err := locker.Lock(ctx, flowID1)
	require.NoError(suite.T(), err)
	defer func() {
		err := unlock1(ctx)
		assert.NoError(suite.T(), err)
	}()

	unlock2, err := locker.Lock(ctx, flowID2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), unlock2)
	defer func() {
		err := unlock2(ctx)
		assert.NoError(suite.T(), err)
	}()
}

func (suite *RedisLockerTestSuite) TestUnlock_ReturnsError_OnMultipleCalls() {
	locker := suite.getTestRedisLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	unlock, err := locker.Lock(ctx, flowID)
	require.NoError(suite.T(), err)

	err = unlock(ctx)
	assert.NoError(suite.T(), err, "first unlock should succeed")

	err = unlock(ctx)
	assert.Error(suite.T(), err, "second unlock should fail")
	assert.Contains(suite.T(), err.Error(), "failed to release lock")

	err = unlock(ctx)
	assert.Error(suite.T(), err, "third unlock should also fail")
}

func (suite *RedisLockerTestSuite) TestConcurrentLockAttempts() {
	locker := suite.getTestRedisLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	numGoroutines := 10
	successCount := 0
	failCount := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			unlock, err := locker.Lock(ctx, flowID)

			mu.Lock()
			if err == nil {
				successCount++
				time.Sleep(10 * time.Millisecond)
				unlockErr := unlock(ctx)
				if unlockErr != nil {
					suite.T().Errorf("unlock failed: %v", unlockErr)
				}
			} else {
				failCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	assert.Equal(suite.T(), 1, successCount, "exactly one lock should succeed")
	assert.Equal(suite.T(), numGoroutines-1, failCount, "all others should fail")
}

func (suite *RedisLockerTestSuite) TestConcurrentDifferentFlows() {
	locker := suite.getTestRedisLocker()
	ctx := context.Background()

	numFlows := 100
	var wg sync.WaitGroup
	wg.Add(numFlows)

	errors := make(chan error, numFlows*2) // Both lock and unlock errors

	for i := 0; i < numFlows; i++ {
		go func() {
			defer wg.Done()

			flowID := uuid.Must(uuid.NewV4())
			unlock, err := locker.Lock(ctx, flowID)

			if err != nil {
				errors <- fmt.Errorf("lock error: %w", err)
				return
			}

			time.Sleep(5 * time.Millisecond)

			if unlockErr := unlock(ctx); unlockErr != nil {
				errors <- fmt.Errorf("unlock error: %w", unlockErr)
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		suite.T().Errorf("unexpected error: %v", err)
	}
}

func (suite *RedisLockerTestSuite) TestLockExpiry() {
	locker := NewRedisLocker(RedisLockerConfig{
		Address: suite.redisAddress,
		Expiry:  1 * time.Second,
	})

	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	_, err := locker.Lock(ctx, flowID)
	require.NoError(suite.T(), err)

	// Unlock ignored, we don't unlock but let it expire
	time.Sleep(2 * time.Second)

	// Should be able to acquire lock again after expiry
	unlock2, err := locker.Lock(ctx, flowID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), unlock2)

	err = unlock2(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *RedisLockerTestSuite) TestContextCancellation() {
	locker := suite.getTestRedisLocker()
	flowID := uuid.Must(uuid.NewV4())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	unlock, err := locker.Lock(ctx, flowID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), unlock)
}

func (suite *RedisLockerTestSuite) TestContextTimeout() {
	locker := suite.getTestRedisLocker()
	flowID := uuid.Must(uuid.NewV4())

	unlock1, err := locker.Lock(context.Background(), flowID)
	require.NoError(suite.T(), err)
	defer func() {
		err := unlock1(context.Background())
		assert.NoError(suite.T(), err)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	unlock2, err := locker.Lock(ctx, flowID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), unlock2)
}

func (suite *RedisLockerTestSuite) TestSimulateRealWorldScenario() {
	locker := suite.getTestRedisLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	results := make(chan bool, 5)
	var wg sync.WaitGroup
	wg.Add(5)

	for i := 0; i < 5; i++ {
		go func(requestNum int) {
			defer wg.Done()

			unlock, err := locker.Lock(ctx, flowID)
			if err != nil {
				results <- false
				return
			}

			time.Sleep(50 * time.Millisecond)

			unlockErr := unlock(ctx)
			if unlockErr != nil {
				suite.T().Errorf("unlock failed for request %d: %v", requestNum, unlockErr)
				results <- false
				return
			}

			results <- true
		}(i)
	}

	wg.Wait()
	close(results)

	successCount := 0
	for success := range results {
		if success {
			successCount++
		}
	}

	assert.Equal(suite.T(), 1, successCount, "only one request should acquire the lock")
}

func (suite *RedisLockerTestSuite) TestMultipleInstances() {
	locker1 := suite.getTestRedisLocker()
	locker2 := NewRedisLocker(RedisLockerConfig{
		Address: suite.redisAddress,
		Expiry:  10 * time.Second,
	})

	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	unlock1, err := locker1.Lock(ctx, flowID)
	require.NoError(suite.T(), err)
	defer func() {
		err := unlock1(ctx)
		assert.NoError(suite.T(), err)
	}()

	unlock2, err := locker2.Lock(ctx, flowID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), unlock2)
}

func (suite *RedisLockerTestSuite) TestRaceCondition() {
	locker := suite.getTestRedisLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	var wg sync.WaitGroup
	iterations := 50

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			unlock, err := locker.Lock(ctx, flowID)
			if err == nil {
				time.Sleep(5 * time.Millisecond)
				if unlockErr := unlock(ctx); unlockErr != nil {
					suite.T().Errorf("unlock failed: %v", unlockErr)
				}
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			unlock, err := locker.Lock(ctx, flowID)
			if err == nil {
				time.Sleep(5 * time.Millisecond)
				if unlockErr := unlock(ctx); unlockErr != nil {
					suite.T().Errorf("unlock failed: %v", unlockErr)
				}
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	wg.Wait()
}

func (suite *RedisLockerTestSuite) TestRedisLocker_UnlockReturnsError() {
	// Test that calling unlock multiple times returns an error
	// This validates that unlock properly returns errors without needing to simulate Redis failures
	locker := suite.getTestRedisLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	unlock, err := locker.Lock(ctx, flowID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), unlock)

	err = unlock(ctx)
	assert.NoError(suite.T(), err, "first unlock should succeed")

	unlockCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = unlock(unlockCtx)
	assert.Error(suite.T(), err, "second unlock should return an error")
	assert.Contains(suite.T(), err.Error(), "failed to release lock")
}
