package webhooks

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"time"
)

type Job struct {
	Data JobData
	Hook Webhook
}

type JobData struct {
	Token string       `json:"token"`
	Event events.Event `json:"event"`
}

type Worker struct {
	logger      echo.Logger
	hookChannel chan Job
}

func NewWorker(hookChannel chan Job, logger echo.Logger) Worker {
	return Worker{
		logger:      logger,
		hookChannel: hookChannel,
	}
}

func (w *Worker) Run() {
	for {
		job, open := <-w.hookChannel
		if !open {
			break
		}

		err := w.triggerWebhook(job)
		if err != nil {
			w.logger.Error(fmt.Errorf("unable to trigger webhook: %w", err))
			continue
		}
	}
}

func (w *Worker) triggerWebhook(job Job) error {
	now := time.Now()
	// check for expire date
	err := job.Hook.DisableOnExpiryDate(now)
	if err != nil {
		return err
	}

	if job.Hook.IsEnabled() {
		err := job.Hook.Trigger(job.Data)
		if err != nil {
			// expire after failure (if failure counter > FailureExpireRate)
			disableErr := job.Hook.DisableOnFailure()
			if disableErr != nil {
				return disableErr
			}

			return err
		}

		err = job.Hook.Reset()
		if err != nil {
			return err
		}
	}

	return nil
}
