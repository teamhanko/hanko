package persistence

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type OTPSecretPersister interface {
	Create(models.OTPSecret) error
	Update(*models.OTPSecret) error
	Delete(*models.OTPSecret) error
}

type otpSecretPersister struct {
	db *pop.Connection
}

func NewOTPSecretPersister(db *pop.Connection) OTPSecretPersister {
	return &otpSecretPersister{db: db}
}

func (p *otpSecretPersister) Create(secret models.OTPSecret) error {
	vErr, err := p.db.ValidateAndCreate(&secret)
	if err != nil {
		return fmt.Errorf("failed to store otp secret credential: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("otp secret object validation failed: %w", vErr)
	}

	return nil
}

func (p *otpSecretPersister) Update(secret *models.OTPSecret) error {
	vErr, err := p.db.ValidateAndUpdate(secret)
	if err != nil {
		return fmt.Errorf("failed to update otp secret: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("otp secret object validation failed: %w", vErr)
	}

	return nil
}

func (p *otpSecretPersister) Delete(secret *models.OTPSecret) error {
	err := p.db.Destroy(secret)
	if err != nil {
		return fmt.Errorf("failed to delete otp secret: %w", err)
	}

	return nil
}
