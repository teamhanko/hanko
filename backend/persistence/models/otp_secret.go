package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

type OTPSecret struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Secret    string    `db:"secret"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (otpSecret OTPSecret) TableName() string {
	return "otp_secrets"
}

func NewOTPSecret(userID uuid.UUID, secret string) *OTPSecret {
	id, _ := uuid.NewV4()
	return &OTPSecret{
		ID:        id,
		UserID:    userID,
		Secret:    secret,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (otpSecret *OTPSecret) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: otpSecret.ID},
		&validators.UUIDIsPresent{Name: "UserID", Field: otpSecret.UserID},
		&validators.StringIsPresent{Name: "Secret", Field: otpSecret.Secret},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: otpSecret.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: otpSecret.UpdatedAt},
	), nil
}
