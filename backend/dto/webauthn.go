package dto

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type WebauthnCredentialUpdateRequest struct {
	Name *string `json:"name"`
}

type WebauthnCredentialResponse struct {
	ID              string     `json:"id"`
	Name            *string    `json:"name,omitempty"`
	PublicKey       string     `json:"public_key"`
	AttestationType string     `json:"attestation_type"`
	AAGUID          uuid.UUID  `json:"aaguid"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	Transports      []string   `json:"transports"`
	BackupEligible  bool       `json:"backup_eligible"`
	BackupState     bool       `json:"backup_state"`
}

// FromWebauthnCredentialModel Converts the DB model to a DTO object
func FromWebauthnCredentialModel(c *models.WebauthnCredential) *WebauthnCredentialResponse {
	return &WebauthnCredentialResponse{
		ID:              c.ID,
		Name:            c.Name,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		AAGUID:          c.AAGUID,
		LastUsedAt:      c.LastUsedAt,
		CreatedAt:       c.CreatedAt,
		Transports:      c.Transports.GetNames(),
		BackupEligible:  c.BackupEligible,
		BackupState:     c.BackupState,
	}
}
