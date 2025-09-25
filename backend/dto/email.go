package dto

import (
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type EmailResponse struct {
	ID         uuid.UUID  `json:"id"`
	Address    string     `json:"address"`
	IsVerified bool       `json:"is_verified"`
	IsPrimary  bool       `json:"is_primary"`
	Identity   *Identity  `json:"identity,omitempty"` // Deprecated
	Identities Identities `json:"identities,omitempty"`
}

type EmailCreateRequest struct {
	Address string `json:"address"`
}

type EmailUpdateRequest struct {
	IsPrimary *bool `json:"is_primary"`
}

// FromEmailModel Converts the DB model to a DTO object
func FromEmailModel(email *models.Email, cfg *config.Config) *EmailResponse {
	emailResponse := &EmailResponse{
		ID:         email.ID,
		Address:    email.Address,
		IsVerified: email.Verified,
		IsPrimary:  email.IsPrimary(),
		Identities: FromIdentitiesModel(email.Identities, cfg),
	}

	if len(email.Identities) > 0 {
		identity := FromIdentityModel(&email.Identities[0], cfg)
		emailResponse.Identity = identity
	}

	return emailResponse
}

type EmailJWT struct {
	Address    string `json:"address"`
	IsPrimary  bool   `json:"is_primary"`
	IsVerified bool   `json:"is_verified"`
}

func (e *EmailJWT) String() string {
	if e == nil {
		return ""
	}
	jsonBytes, _ := json.Marshal(e)
	return string(jsonBytes)
}

func EmailJWTFromEmailModel(email *models.Email) *EmailJWT {
	if email == nil {
		return nil
	}

	return &EmailJWT{
		Address:    email.Address,
		IsPrimary:  email.IsPrimary(),
		IsVerified: email.Verified,
	}
}
