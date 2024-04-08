package dto

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
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
func FromEmailModel(email *models.Email) *EmailResponse {
	emailResponse := &EmailResponse{
		ID:         email.ID,
		Address:    email.Address,
		IsVerified: email.Verified,
		IsPrimary:  email.IsPrimary(),
		Identities: FromIdentitiesModel(email.Identities),
	}

	if len(email.Identities) > 0 {
		identity := FromIdentityModel(&email.Identities[0])
		emailResponse.Identity = identity
	}

	return emailResponse
}

type EmailJwt struct {
	Address    string `json:"address"`
	IsPrimary  bool   `json:"is_primary"`
	IsVerified bool   `json:"is_verified"`
}

func JwtFromEmailModel(email *models.Email) *EmailJwt {
	if email == nil {
		return nil
	}

	return &EmailJwt{
		Address:    email.Address,
		IsPrimary:  email.IsPrimary(),
		IsVerified: email.Verified,
	}
}
