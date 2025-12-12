package models

import (
	"time"

	"github.com/gofrs/uuid"
)

type SecurityNotification struct {
	ID           uuid.UUID `db:"id"`
	EmailAddress string    `db:"email_address"`
	TemplateName string    `db:"template_name"`
	Language     string    `db:"language"`
	CreatedAt    time.Time `db:"created_at"`
	Email        Email     `belongs_to:"email"`
}
