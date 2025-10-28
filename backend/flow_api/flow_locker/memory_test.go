package flow_locker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryLocker_Lock_Success(t *testing.T) {
	locker := NewMemoryLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	unlock, err := locker.Lock(ctx, flowID)

	require.NoError(t, err)
	require.NotNil(t, unlock)

	// Clean up
	err = unlock(ctx)
	assert.NoError(t, err)
}

func TestMemoryLocker_Lock_FailFast_WhenAlreadyLocked(t *testing.T) {
	locker := NewMemoryLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	// First lock succeeds
	unlock1, err := locker.Lock(ctx, flowID)
	require.NoError(t, err)
	require.NotNil(t, unlock1)

	// Second lock on same flow ID should fail immediately
	unlock2, err := locker.Lock(ctx, flowID)

	assert.Error(t, err)
	assert.Nil(t, unlock2)
	assert.Contains(t, err.Error(), "already being processed")

	// Clean up first lock
	err = unlock1(ctx)
	assert.NoError(t, err)
}

func TestMemoryLocker_Lock_SucceedsAfterUnlock(t *testing.T) {
	locker := NewMemoryLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	// First lock
	unlock1, err := locker.Lock(ctx, flowID)
	require.NoError(t, err)
	err = unlock1(ctx)
	require.NoError(t, err)

	// Second lock should now succeed
	unlock2, err := locker.Lock(ctx, flowID)
	require.NoError(t, err)
	require.NotNil(t, unlock2)
	err = unlock2(ctx)
	assert.NoError(t, err)
}

func TestMemoryLocker_Lock_DifferentFlowIDs(t *testing.T) {
	locker := NewMemoryLocker()
	ctx := context.Background()
	flowID1 := uuid.Must(uuid.NewV4())
	flowID2 := uuid.Must(uuid.NewV4())

	// Lock first flow
	unlock1, err := locker.Lock(ctx, flowID1)
	require.NoError(t, err)

	// Lock second flow should succeed (different ID)
	unlock2, err := locker.Lock(ctx, flowID2)
	require.NoError(t, err)
	require.NotNil(t, unlock2)

	// Clean up
	err = unlock2(ctx)
	assert.NoError(t, err)
	err = unlock1(ctx)
	assert.NoError(t, err)
}

func TestMemoryLocker_Unlock_CanBeCalled_Multiple(t *testing.T) {
	locker := NewMemoryLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	unlock, err := locker.Lock(ctx, flowID)
	require.NoError(t, err)

	// First unlock should succeed
	err = unlock(ctx)
	assert.NoError(t, err)

	// Subsequent unlocks should not panic (even if they're no-ops)
	assert.NotPanics(t, func() {
		unlock(ctx)
		unlock(ctx)
	})
}

func TestMemoryLocker_ConcurrentLockAttempts(t *testing.T) {
	locker := NewMemoryLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	numGoroutines := 10
	successCount := 0
	failCount := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(numGoroutines)

	// Launch multiple goroutines trying to lock the same flow
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			unlock, err := locker.Lock(ctx, flowID)

			mu.Lock()
			if err == nil {
				successCount++
				// Hold lock briefly
				time.Sleep(10 * time.Millisecond)
				unlockErr := unlock(ctx)
				if unlockErr != nil {
					t.Errorf("unlock failed: %v", unlockErr)
				}
			} else {
				failCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Exactly one should succeed, rest should fail
	assert.Equal(t, 1, successCount, "exactly one lock should succeed")
	assert.Equal(t, numGoroutines-1, failCount, "all others should fail")
}

func TestMemoryLocker_ConcurrentDifferentFlows(t *testing.T) {
	locker := NewMemoryLocker()
	ctx := context.Background()

	numFlows := 100
	var wg sync.WaitGroup
	wg.Add(numFlows)

	errors := make(chan error, numFlows*2) // Both lock and unlock errors

	// Lock different flows concurrently
	for i := 0; i < numFlows; i++ {
		go func() {
			defer wg.Done()

			flowID := uuid.Must(uuid.NewV4())
			unlock, err := locker.Lock(ctx, flowID)

			if err != nil {
				errors <- err
				return
			}

			// Simulate work
			time.Sleep(5 * time.Millisecond)

			if unlockErr := unlock(ctx); unlockErr != nil {
				errors <- unlockErr
			}
		}()
	}

	wg.Wait()
	close(errors)

	// All should succeed since they're different flow IDs
	for err := range errors {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMemoryLocker_MemoryLeak_LocksAreCleanedUp(t *testing.T) {
	locker := NewMemoryLocker()
	ctx := context.Background()

	// Lock and unlock many flows
	for i := 0; i < 1000; i++ {
		flowID := uuid.Must(uuid.NewV4())
		unlock, err := locker.Lock(ctx, flowID)
		require.NoError(t, err)
		err = unlock(ctx)
		require.NoError(t, err)
	}

	// Check that the map is cleaned up
	locker.mu.Lock()
	mapSize := len(locker.locks)
	locker.mu.Unlock()

	assert.Equal(t, 0, mapSize, "locks map should be empty after all unlocks")
}

func TestMemoryLocker_RaceCondition(t *testing.T) {
	// This test is designed to catch race conditions when run with -race flag
	locker := NewMemoryLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	var wg sync.WaitGroup
	iterations := 100

	// Goroutine 1: Repeatedly lock and unlock
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			unlock, err := locker.Lock(ctx, flowID)
			if err == nil {
				time.Sleep(time.Microsecond)
				if unlockErr := unlock(ctx); unlockErr != nil {
					t.Errorf("unlock failed: %v", unlockErr)
				}
			}
			time.Sleep(time.Microsecond)
		}
	}()

	// Goroutine 2: Repeatedly try to lock
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			unlock, err := locker.Lock(ctx, flowID)
			if err == nil {
				time.Sleep(time.Microsecond)
				if unlockErr := unlock(ctx); unlockErr != nil {
					t.Errorf("unlock failed: %v", unlockErr)
				}
			}
			time.Sleep(time.Microsecond)
		}
	}()

	wg.Wait()
}

func TestMemoryLocker_ContextCancellation_DoesNotAffectLocking(t *testing.T) {
	locker := NewMemoryLocker()
	flowID := uuid.Must(uuid.NewV4())

	// Create a context (not canceled yet)
	ctx, cancel := context.WithCancel(context.Background())

	// Acquire lock with VALID context - this succeeds
	unlock, err := locker.Lock(ctx, flowID)
	require.NoError(t, err)
	require.NotNil(t, unlock)

	// Cancel the context AFTER lock is acquired
	cancel()

	// Try to acquire the SAME lock from a NEW context,
	// verify lock is still held
	_, err = locker.Lock(context.Background(), flowID)
	assert.Error(t, err)

	// Clean up - unlock should still work with original (canceled) context
	err = unlock(ctx)
	assert.NoError(t, err)
}

func TestMemoryLocker_SimulateRealWorldScenario(t *testing.T) {
	// Simulate 5 parallel requests for the same flow (like your k6 test)
	locker := NewMemoryLocker()
	ctx := context.Background()
	flowID := uuid.Must(uuid.NewV4())

	results := make(chan bool, 5)
	var wg sync.WaitGroup
	wg.Add(5)

	// Simulate 5 parallel requests
	for i := 0; i < 5; i++ {
		go func(requestNum int) {
			defer wg.Done()

			unlock, err := locker.Lock(ctx, flowID)
			if err != nil {
				// Request failed to acquire lock
				results <- false
				return
			}

			// Simulate processing (e.g., verifying passcode)
			time.Sleep(50 * time.Millisecond)

			unlockErr := unlock(ctx)
			if unlockErr != nil {
				t.Errorf("unlock failed for request %d: %v", requestNum, unlockErr)
				results <- false
				return
			}

			results <- true
		}(i)
	}

	wg.Wait()
	close(results)

	// Count successes
	successCount := 0
	for success := range results {
		if success {
			successCount++
		}
	}

	// Only 1 request should have succeeded
	assert.Equal(t, 1, successCount, "only one request should acquire the lock")
}
