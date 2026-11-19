package persistence

import (
	"errors"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type OTPSecretPersister interface {
	GetForUpdate(id uuid.UUID, tenantID uuid.UUID) (*models.OTPSecret, error)
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

// GetForUpdate locks the secret until the surrounding transaction ends, so it must be called inside one and
// returns an error otherwise.
func (p *otpSecretPersister) GetForUpdate(id uuid.UUID, tenantID uuid.UUID) (*models.OTPSecret, error) {
	if p.db.TX == nil {
		return nil, errors.New("GetForUpdate must be called inside a transaction")
	}

	secret := models.OTPSecret{}

	err := p.db.RawQuery("SELECT * FROM otp_secrets WHERE id = ? AND tenant_id = ? FOR UPDATE", id, tenantID).First(&secret)
	if err != nil {
		return nil, fmt.Errorf("failed to get otp secret for update: %w", err)
	}

	return &secret, nil
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
