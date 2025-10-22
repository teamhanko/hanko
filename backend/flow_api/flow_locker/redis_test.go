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

// RedisLockerTestSuite holds the test suite with shared Redis setup
type RedisLockerTestSuite struct {
	suite.Suite
	pool         *dockertest.Pool
	resource     *dockertest.Resource
	redisAddress string
}

// SetupSuite runs once before all tests in the suite
func (suite *RedisLockerTestSuite) SetupSuite() {
	var err error

	// Create dockertest pool
	suite.pool, err = dockertest.NewPool("")
	require.NoError(suite.T(), err, "Could not construct pool")

	err = suite.pool.Client.Ping()
	require.NoError(suite.T(), err, "Could not connect to Docker")

	// Pull and run Redis container
	suite.resource, err = suite.pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "redis",
		Tag:        "8-alpine",
		Env:        []string{},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(suite.T(), err, "Could not start Redis container")

	// Set container to expire in 10 minutes
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
		unlock()
		return nil
	})
	require.NoError(suite.T(), err, "Could not connect to Redis")
}

// TearDownSuite runs once after all tests in the suite
func (suite *RedisLockerTestSuite) TearDownSuite() {
	if suite.resource != nil {
		err := suite.pool.Purge(suite.resource)
		require.NoError(suite.T(), err, "Could not purge Redis container")
	}
}

// getTestRedisLocker is a helper to create a locker for each test
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

// Tests start here

func (suite *RedisLockerTestSuite) TestLock_Success() {
	locker := suite.getTestRedisLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	unlock, err := locker.Lock(ctx, flowID)

	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), unlock)

	unlock()
}

func (suite *RedisLockerTestSuite) TestLock_FailFast_WhenAlreadyLocked() {
	locker := suite.getTestRedisLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	unlock1, err := locker.Lock(ctx, flowID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), unlock1)
	defer unlock1()

	unlock2, err := locker.Lock(ctx, flowID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), unlock2)
	assert.Contains(suite.T(), err.Error(), "failed to acquire lock")
}

func (suite *RedisLockerTestSuite) TestLock_SucceedsAfterUnlock() {
	locker := suite.getTestRedisLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	unlock1, err := locker.Lock(ctx, flowID)
	require.NoError(suite.T(), err)
	unlock1()

	time.Sleep(10 * time.Millisecond)

	unlock2, err := locker.Lock(ctx, flowID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), unlock2)
	unlock2()
}

func (suite *RedisLockerTestSuite) TestLock_DifferentFlowIDs() {
	locker := suite.getTestRedisLocker()
	ctx := context.Background()
	flowID1 := uuid.Must(uuid.NewV4())
	flowID2 := uuid.Must(uuid.NewV4())

	unlock1, err := locker.Lock(ctx, flowID1)
	require.NoError(suite.T(), err)
	defer unlock1()

	unlock2, err := locker.Lock(ctx, flowID2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), unlock2)
	defer unlock2()
}

func (suite *RedisLockerTestSuite) TestUnlock_CanBeCalled_Multiple() {
	locker := suite.getTestRedisLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	unlock, err := locker.Lock(ctx, flowID)
	require.NoError(suite.T(), err)

	assert.NotPanics(suite.T(), func() {
		unlock()
		unlock()
		unlock()
	})
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
				unlock()
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

	errors := make(chan error, numFlows)

	for i := 0; i < numFlows; i++ {
		go func() {
			defer wg.Done()

			flowID := uuid.Must(uuid.NewV4())
			unlock, err := locker.Lock(ctx, flowID)

			if err != nil {
				errors <- err
				return
			}

			time.Sleep(5 * time.Millisecond)
			unlock()
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

	unlock1, err := locker.Lock(ctx, flowID)
	require.NoError(suite.T(), err)
	defer unlock1()

	time.Sleep(2 * time.Second)

	unlock2, err := locker.Lock(ctx, flowID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), unlock2)
	unlock2()
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
	defer unlock1()

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
			unlock()

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
	defer unlock1()

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
				unlock()
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
				unlock()
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	wg.Wait()
}
