package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

type TrustedDevice struct {
	ID          uuid.UUID `db:"id"`
	UserID      uuid.UUID `db:"user_id"`
	DeviceToken string    `db:"device_token"`
	ExpiresAt   time.Time `db:"expires_at"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (trustedDevice *TrustedDevice) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: trustedDevice.ID},
		&validators.UUIDIsPresent{Name: "UserID", Field: trustedDevice.UserID},
		&validators.StringIsPresent{Name: "DeviceToken", Field: trustedDevice.DeviceToken},
		&validators.TimeIsPresent{Name: "ExpiresAt", Field: trustedDevice.ExpiresAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: trustedDevice.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: trustedDevice.CreatedAt},
	), nil
}
