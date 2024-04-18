package events

import "github.com/teamhanko/hanko/backend/persistence/models"

type Event string

const (
	User             Event = "user"
	UserCreate       Event = "user.create"
	UserUpdate       Event = "user.update"
	UserDelete       Event = "user.delete"
	UserEmail        Event = "user.update.email"
	UserEmailCreate  Event = "user.update.email.create"
	UserEmailPrimary Event = "user.update.email.primary"
	UserEmailDelete  Event = "user.update.email.delete"

	EmailSend Event = "email.send"
)

func StringIsValidEvent(value string) bool {
	evt := Event(value)
	return IsValidEvent(evt)
}

func IsValidEvent(evt Event) bool {
	var isValid bool
	switch evt {
	case User, UserCreate, UserUpdate, UserDelete, UserEmail, UserEmailCreate, UserEmailPrimary, UserEmailDelete, EmailSend:
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
