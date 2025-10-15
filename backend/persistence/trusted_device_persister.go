package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type TrustedDevicePersister interface {
	Create(models.TrustedDevice) error
	FindByDeviceToken(string) (*models.TrustedDevice, error)
}

type trustedDevicePersister struct {
	db *pop.Connection
}

func NewTrustedDevicePersister(db *pop.Connection) TrustedDevicePersister {
	return &trustedDevicePersister{db: db}
}

func (p *trustedDevicePersister) Create(trustedDevice models.TrustedDevice) error {
	vErr, err := p.db.ValidateAndCreate(&trustedDevice)
	if err != nil {
		return fmt.Errorf("failed to store trustedDevice: %w", err)
	}
	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("trustedDevice object validation failed: %w", vErr)
	}

	return nil
}

func (p *trustedDevicePersister) FindByDeviceToken(token string) (*models.TrustedDevice, error) {
	trustedDevice := models.TrustedDevice{}
	err := p.db.Where("device_token = ?", token).First(&trustedDevice)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get trustedDevice: %w", err)
	}

	return &trustedDevice, nil
}
