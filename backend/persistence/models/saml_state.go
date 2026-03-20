package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

type SamlState struct {
	ID        uuid.UUID  `db:"id"`
	Nonce     string     `db:"nonce"`
	State     string     `db:"state"`
	ExpiresAt time.Time  `db:"expires_at"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	TenantID  *uuid.UUID `db:"tenant_id"`
}

func NewSamlState(nonce string, state string, tenantID *uuid.UUID) (*SamlState, error) {
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
		TenantID:  tenantID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
