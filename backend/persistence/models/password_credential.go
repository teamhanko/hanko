package models

import (
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

type PasswordCredential struct {
	ID        uuid.UUID  `db:"id"`
	UserId    uuid.UUID  `db:"user_id"`
	TenantID  *uuid.UUID `db:"tenant_id"`
	Password  string     `db:"password"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

func NewPasswordCredential(userId uuid.UUID, password string, tenantID *uuid.UUID) *PasswordCredential {
	id, _ := uuid.NewV4()
	return &PasswordCredential{
		ID:        id,
		UserId:    userId,
		Password:  password,
		TenantID:  tenantID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

func (password *PasswordCredential) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Name: "Password", Field: password.Password},
		&validators.UUIDIsPresent{Name: "UserId", Field: password.UserId},
	), nil
}
