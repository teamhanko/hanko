package models

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/crypto"
	"time"
)

type Session struct {
	ID        string    `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewSession(userID uuid.UUID) (*Session, error) {
	if userID.IsNil() {
		return nil, errors.New("userID is required")
	}

	now := time.Now().UTC()

	// Session Key (according to OWASP at least 16 bytes)
	key, err := crypto.GenerateRandomStringURLSafe(32)
	if err != nil {
		return nil, fmt.Errorf("could not generate random bytes: %w", err)
	}

	return &Session{
		ID:        key,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (session *Session) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Name: "ID", Field: session.ID},
		&validators.UUIDIsPresent{Name: "UserID", Field: session.UserID},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: session.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: session.CreatedAt},
	), nil
}
