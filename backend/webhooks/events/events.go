package events

import "github.com/teamhanko/hanko/backend/persistence/models"

type Event string

const (
	UserCreate   Event = "user.create"
	UserUpdate   Event = "user.update"
	UserDelete   Event = "user.delete"
	EmailCreate  Event = "user.update.email.create"
	EmailPrimary Event = "user.update.email.primary"
	EmailDelete  Event = "user.update.email.delete"
)

func StringIsValidEvent(value string) bool {
	evt := Event(value)
	return IsValidEvent(evt)
}

func IsValidEvent(evt Event) bool {
	var isValid bool
	switch evt {
	case "user", "user.update.email", UserCreate, UserUpdate, UserDelete, EmailCreate, EmailPrimary, EmailDelete:
		isValid = true
	default:
		isValid = false
	}

	return isValid
}

type Events []Event

func ConvertFromDbList(events models.WebhookEvents) Events {
	evts := make(Events, 0)
	for _, event := range events {
		evts = append(evts, Event(event.Event))
	}

	return evts
}
