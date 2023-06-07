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

type Token struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Value     string    `db:"value"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewToken(userID uuid.UUID) (*Token, error) {
	if userID.IsNil() {
		return nil, errors.New("userID is required")
	}

	now := time.Now().UTC()

	id, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("could not generate id: %w", err)
	}

	value, err := crypto.GenerateRandomStringURLSafe(32)
	if err != nil {
		return nil, fmt.Errorf("could not generate random string: %w", err)
	}

	return &Token{
		ID:        id,
		UserID:    userID,
		Value:     value,
		ExpiresAt: now.Add(time.Minute),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (token *Token) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: token.ID},
		&validators.UUIDIsPresent{Name: "UserID", Field: token.UserID},
		&validators.StringIsPresent{Name: "Value", Field: token.Value},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: token.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: token.CreatedAt},
	), nil
}
