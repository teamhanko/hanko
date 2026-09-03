package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type RoleBindingPersister interface {
	Get(userID uuid.UUID, roleID uuid.UUID, organizationID uuid.UUID, tenantID uuid.UUID) (*models.RoleBinding, error)
	Create(binding models.RoleBinding) error
	Delete(binding models.RoleBinding) error
	ListByUserAndOrganization(userID uuid.UUID, organizationID uuid.UUID, tenantID uuid.UUID) ([]models.RoleBinding, error)
}

type roleBindingPersister struct {
	db *pop.Connection
}

func NewRoleBindingPersister(db *pop.Connection) RoleBindingPersister {
	return &roleBindingPersister{db: db}
}

func (p *roleBindingPersister) Get(userID uuid.UUID, roleID uuid.UUID, organizationID uuid.UUID, tenantID uuid.UUID) (*models.RoleBinding, error) {
	binding := models.RoleBinding{}
	err := p.db.
		Where("tenant_id = ?", tenantID).
		Where("user_id = ?", userID).
		Where("role_id = ?", roleID).
		Where("organization_id = ?", organizationID).
		First(&binding)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role binding: %w", err)
	}

	return &binding, nil
}

func (p *roleBindingPersister) Create(binding models.RoleBinding) error {
	vErr, err := p.db.ValidateAndCreate(&binding)
	if err != nil {
		return fmt.Errorf("failed to store role binding: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("role binding object validation failed: %w", vErr)
	}

	return nil
}

func (p *roleBindingPersister) Delete(binding models.RoleBinding) error {
	err := p.db.Destroy(&binding)
	if err != nil {
		return fmt.Errorf("failed to delete role binding: %w", err)
	}

	return nil
}

func (p *roleBindingPersister) ListByUserAndOrganization(userID uuid.UUID, organizationID uuid.UUID, tenantID uuid.UUID) ([]models.RoleBinding, error) {
	bindings := []models.RoleBinding{}

	err := p.db.
		Where("tenant_id = ?", tenantID).
		Where("user_id = ?", userID).
		Where("organization_id = ?", organizationID).
		Order("created_at desc").
		All(&bindings)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return bindings, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch role bindings: %w", err)
	}

	return bindings, nil
}
