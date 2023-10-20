package models

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"strings"
	"time"
)

type SamlState struct {
	ID        uuid.UUID `db:"id"`
	Nonce     string    `db:"nonce"`
	State     string    `db:"state"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewSamlState(nonce string, state string) (*SamlState, error) {
	if strings.TrimSpace(nonce) == "" {
		return nil, errors.New("nonce is required")
	}

	if strings.TrimSpace(state) == "" {
		return nil, errors.New("state is required")
	}

	now := time.Now().UTC()

	id, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("could not generate id: %w", err)
	}

	return &SamlState{
		ID:        id,
		Nonce:     nonce,
		State:     state,
		ExpiresAt: now.Add(time.Minute),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
