package webhooks

import (
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
	hookChannel chan Job
}

func NewWorker(hookChannel chan Job) Worker {
	return Worker{
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

	if !job.Hook.IsEnabled() {
		err := job.Hook.Trigger(job.Data)
		if err != nil {
			// expire after failure (if failure counter > FailureExpireRate)
			err := job.Hook.DisableOnFailure()
			if err != nil {
				return err
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
