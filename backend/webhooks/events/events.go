package events

import "github.com/teamhanko/hanko/backend/v2/persistence/models"

type Event string

const (
	User               Event = "user"
	UserLogin          Event = "user.login"
	UserCreate         Event = "user.create"
	UserUpdate         Event = "user.update"
	UserDelete         Event = "user.delete"
	UserEmail          Event = "user.update.email"
	UserEmailCreate    Event = "user.update.email.create"
	UserEmailPrimary   Event = "user.update.email.primary"
	UserEmailDelete    Event = "user.update.email.delete"
	UserUsername       Event = "user.update.username"
	UserUsernameCreate Event = "user.update.username.create"
	UserUsernameDelete Event = "user.update.username.delete"
	UserUsernameUpdate Event = "user.update.username.update"
	UserPasswordChange Event = "user.update.password.update"

	EmailSend Event = "email.send"

	Tenant       Event = "tenant"
	TenantCreate Event = "tenant.create"
	TenantUpdate Event = "tenant.update"
	TenantDelete Event = "tenant.delete"
)

func StringIsValidEvent(value string) bool {
	evt := Event(value)
	return IsValidEvent(evt)
}

func IsValidEvent(evt Event) bool {
	var isValid bool
	switch evt {
	case User, UserLogin, UserCreate, UserUpdate, UserDelete, UserEmail, UserEmailCreate, UserEmailPrimary, UserEmailDelete, UserUsername, UserUsernameCreate, UserUsernameUpdate, UserUsernameDelete, UserPasswordChange, EmailSend, Tenant, TenantCreate, TenantUpdate, TenantDelete:
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
