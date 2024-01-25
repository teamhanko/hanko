package webhooks

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"testing"
	"time"
)

type TestLogger struct {
	log.Logger
	ErrorLogFunc func()
}

func (tl *TestLogger) Error(_ ...interface{}) {
	tl.ErrorLogFunc()
}

type TestWorker struct {
	Worker
}

type TestHook struct {
	ExpireFunc    func() error
	FailureFunc   func() error
	IsEnabledFunc func() bool
	ResetFunc     func() error
	TriggerFunc   func() error
}

func (th *TestHook) DisableOnExpiryDate(_ time.Time) error {
	return th.ExpireFunc()
}

func (th *TestHook) IsEnabled() bool {
	return th.IsEnabledFunc()
}

func (th *TestHook) Trigger(_ JobData) error {
	return th.TriggerFunc()
}

func (th *TestHook) DisableOnFailure() error {
	return th.FailureFunc()
}

func (th *TestHook) Reset() error {
	return th.ResetFunc()
}

func (th *TestHook) HasEvent(_ events.Event) bool {
	return true
}

func TestWorker_RunWithNothing(t *testing.T) {
	hookChannel := make(chan Job)

	worker := TestWorker{NewWorker(hookChannel, log.New("test"))}
	close(hookChannel)
	worker.Run()
}

func TestWorker_RunJob(t *testing.T) {
	job := Job{
		Data: JobData{
			Token: "test-token",
			Event: events.UserCreate,
		},

		Hook: &TestHook{
			ExpireFunc: func() error {
				return nil
			},
			FailureFunc: func() error {
				return nil
			},
			IsEnabledFunc: func() bool {
				return true
			},
			ResetFunc: func() error {
				return nil
			},
			TriggerFunc: func() error {
				require.True(t, true)
				return nil
			},
		},
	}

	hookChannel := make(chan Job, 1)
	hookChannel <- job
	close(hookChannel)

	worker := TestWorker{NewWorker(hookChannel, log.New("test"))}
	worker.Run()
}

func TestWorker_RunJobWithError(t *testing.T) {
	job := Job{
		Data: JobData{
			Token: "test-token",
			Event: events.UserCreate,
		},

		Hook: &TestHook{
			ExpireFunc: func() error {
				return fmt.Errorf("forced error")
			},
		},
	}

	hookChannel := make(chan Job, 1)
	hookChannel <- job
	close(hookChannel)

	worker := TestWorker{NewWorker(hookChannel, &TestLogger{
		Logger:       *log.New("test"),
		ErrorLogFunc: func() { require.True(t, true) },
	})}
	worker.Run()
}

func TestWorker_TriggerWebhook(t *testing.T) {
	job := Job{
		Data: JobData{
			Token: "test-token",
			Event: events.UserCreate,
		},

		Hook: &TestHook{
			ExpireFunc: func() error {
				require.True(t, true)
				return nil
			},
			FailureFunc: func() error {
				require.True(t, true)
				return nil
			},
			IsEnabledFunc: func() bool {
				require.True(t, true)
				return true
			},
			ResetFunc: func() error {
				require.True(t, true)
				return nil
			},
			TriggerFunc: func() error {
				require.True(t, true)
				return nil
			},
		},
	}

	worker := TestWorker{NewWorker(nil, log.New("test"))}
	err := worker.triggerWebhook(job)
	require.NoError(t, err)
}

func TestWorker_TriggerWebhookWithExpireError(t *testing.T) {
	job := Job{
		Data: JobData{
			Token: "test-token",
			Event: events.UserCreate,
		},

		Hook: &TestHook{
			ExpireFunc: func() error {
				require.True(t, true)
				return fmt.Errorf("expired error")
			},
		},
	}

	worker := TestWorker{NewWorker(nil, log.New("test"))}
	err := worker.triggerWebhook(job)
	require.ErrorContains(t, err, "expired error")
}

func TestWorker_TriggerWebhookIgnoreDisabledJob(t *testing.T) {
	job := Job{
		Data: JobData{
			Token: "test-token",
			Event: events.UserCreate,
		},

		Hook: &TestHook{
			ExpireFunc: func() error {
				require.True(t, true)
				return nil
			},
			IsEnabledFunc: func() bool {
				require.True(t, true)
				return false
			},
		},
	}

	worker := TestWorker{NewWorker(nil, log.New("test"))}
	err := worker.triggerWebhook(job)
	require.NoError(t, err)
}

func TestWorker_TriggerWebhookTriggerWithError(t *testing.T) {
	job := Job{
		Data: JobData{
			Token: "test-token",
			Event: events.UserCreate,
		},

		Hook: &TestHook{
			ExpireFunc: func() error {
				require.True(t, true)
				return nil
			},
			IsEnabledFunc: func() bool {
				require.True(t, true)
				return true
			},
			TriggerFunc: func() error {
				require.True(t, true)
				return fmt.Errorf("trigger error")
			},
			FailureFunc: func() error {
				require.True(t, true)
				return nil
			},
		},
	}

	worker := TestWorker{NewWorker(nil, log.New("test"))}
	err := worker.triggerWebhook(job)
	require.ErrorContains(t, err, "trigger error")
}

func TestWorker_TriggerWebhookDisableOnFailure(t *testing.T) {
	job := Job{
		Data: JobData{
			Token: "test-token",
			Event: events.UserCreate,
		},

		Hook: &TestHook{
			ExpireFunc: func() error {
				require.True(t, true)
				return nil
			},
			IsEnabledFunc: func() bool {
				require.True(t, true)
				return true
			},
			TriggerFunc: func() error {
				require.True(t, true)
				return fmt.Errorf("trigger error")
			},
			FailureFunc: func() error {
				require.True(t, true)
				return fmt.Errorf("failure error")
			},
		},
	}

	worker := TestWorker{NewWorker(nil, log.New("test"))}
	err := worker.triggerWebhook(job)
	require.ErrorContains(t, err, "failure error")
}
func TestWorker_TriggerWebhookResetError(t *testing.T) {
	job := Job{
		Data: JobData{
			Token: "test-token",
			Event: events.UserCreate,
		},

		Hook: &TestHook{
			ExpireFunc: func() error {
				require.True(t, true)
				return nil
			},
			IsEnabledFunc: func() bool {
				require.True(t, true)
				return true
			},
			TriggerFunc: func() error {
				require.True(t, true)
				return nil
			},
			FailureFunc: func() error {
				require.True(t, true)
				return nil
			},
			ResetFunc: func() error {
				require.True(t, true)
				return fmt.Errorf("disable error")
			},
		},
	}

	worker := TestWorker{NewWorker(nil, log.New("test"))}
	err := worker.triggerWebhook(job)
	require.ErrorContains(t, err, "disable error")
}
