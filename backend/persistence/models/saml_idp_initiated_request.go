package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

type SamlIDPInitiatedRequest struct {
	ID         uuid.UUID `db:"id"`
	ResponseID string    `db:"response_id"`
	Issuer     string    `db:"issuer"`
	ExpiresAt  time.Time `db:"expires_at"`
	CreatedAt  time.Time `db:"created_at"`
}

func NewSamlIDPInitiatedRequest(responseID, issuer string, expiresAt time.Time) (*SamlIDPInitiatedRequest, error) {
	id, _ := uuid.NewV4()

	return &SamlIDPInitiatedRequest{
		ID:         id,
		ResponseID: responseID,
		Issuer:     issuer,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now().UTC(),
	}, nil
}

func (samlIDPInitiatedRequest SamlIDPInitiatedRequest) TableName() string {
	return "saml_idp_initiated_requests"
}

func (r *SamlIDPInitiatedRequest) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: r.ID},
		&validators.StringIsPresent{Name: "ResponseID", Field: r.ResponseID},
		&validators.StringIsPresent{Name: "Issuer", Field: r.Issuer},
		&validators.TimeIsPresent{Name: "ExpiresAt", Field: r.ExpiresAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: r.CreatedAt},
	), nil
}
