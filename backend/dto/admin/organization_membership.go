package admin

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

// OrganizationMember is a lightweight view of a user's membership in an
// organization - membership metadata only, not the full user record. A
// caller that needs full user detail calls GET /users/{id} separately.
type OrganizationMember struct {
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// FromOrganizationMembershipModel converts the DB model to a DTO object
func FromOrganizationMembershipModel(model models.OrganizationMembership) OrganizationMember {
	return OrganizationMember{
		UserID:    model.UserID,
		CreatedAt: model.CreatedAt,
	}
}
