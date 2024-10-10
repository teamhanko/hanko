package models

import (
	"github.com/gofrs/uuid"
	"time"
)

type Session struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	UserAgent string     `db:"user_agent"`
	IpAddress string     `db:"ip_address"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	ExpiresAt *time.Time `db:"expires_at"`
	LastUsed  time.Time  `db:"last_used"`
}
