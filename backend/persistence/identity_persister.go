package persistence

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type IdentityPersister interface {
	Get(providerUserID string, providerID string, tenantID uuid.UUID) (*models.Identity, error)
	GetByID(identityID uuid.UUID, tenantID uuid.UUID) (*models.Identity, error)
	GetAllByDomain(tenantID uuid.UUID, domain string) ([]models.Identity, error)
	Create(identity models.Identity) error
	Update(identity models.Identity) error
	Delete(identity models.Identity) error
	DeleteAll(identities []models.Identity) error
}

type identityPersister struct {
	db *pop.Connection
}

func (p identityPersister) GetByID(identityID uuid.UUID, tenantID uuid.UUID) (*models.Identity, error) {
	identity := &models.Identity{}
	query := p.db.EagerPreload("Email", "Email.User", "Email.User.Username", "SamlIdentity")
	query = query.Where("identities.tenant_id = ?", tenantID)
	if err := query.Find(identity, identityID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}

	var samlProvider models.SamlProvider
	if identity.SamlIdentity != nil {
		q2 := p.db.RawQuery("select * from saml_providers where tenant_id = $1 and domain = $2", tenantID, identity.SamlIdentity.Domain)
		if err := q2.First(&samlProvider); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, fmt.Errorf("failed to get samlProvider: %w", err)
		}
		identity.SamlIdentity.SamlProvider = &samlProvider
	}
	return identity, nil
}

func (p identityPersister) Get(providerUserID string, providerID string, tenantID uuid.UUID) (*models.Identity, error) {
	identity := &models.Identity{}
	query := p.db.EagerPreload().Where("provider_user_id = ? AND provider_id = ?", providerUserID, providerID)
	query = query.Where("tenant_id = ?", tenantID)
	if err := query.First(identity); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}

	var samlProvider models.SamlProvider
	if identity.SamlIdentity != nil {
		q2 := p.db.RawQuery("select * from saml_providers where tenant_id = $1 and domain = $2", tenantID.String(), identity.SamlIdentity.Domain)
		if err := q2.First(&samlProvider); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, fmt.Errorf("failed to get samlProvider: %w", err)
		}
		identity.SamlIdentity.SamlProvider = &samlProvider
	}
	return identity, nil
}

func (p identityPersister) GetAllByDomain(tenantID uuid.UUID, domain string) ([]models.Identity, error) {
	identities := []models.Identity{}
	query := p.db.EagerPreload("SamlIdentity").
		Join("saml_identities", "saml_identities.identity_id = identities.id").
		Where("identities.tenant_id = ?", tenantID).
		Where("saml_identities.domain = ?", strings.TrimSpace(domain))
	if err := query.All(&identities); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get identities: %w", err)
	}

	return identities, nil
}

func (p identityPersister) Create(identity models.Identity) error {
	vErr, err := p.db.ValidateAndCreate(&identity)
	if err != nil {
		return fmt.Errorf("failed to store identity: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("identity object validation failed: %w", vErr)
	}

	return nil
}

func (p identityPersister) Update(identity models.Identity) error {
	vErr, err := p.db.ValidateAndUpdate(&identity)
	if err != nil {
		return fmt.Errorf("failed to update identity: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("identity object validation failed: %w", vErr)
	}

	return nil
}

func (p identityPersister) Delete(identity models.Identity) error {
	err := p.db.Destroy(&identity)
	if err != nil {
		return fmt.Errorf("failed to delete identity: %w", err)
	}

	return nil
}

func (p identityPersister) DeleteAll(identities []models.Identity) error {
	err := p.db.Q().Delete(identities)
	if err != nil {
		return fmt.Errorf("failed to delete identities: %w", err)
	}

	return nil
}

func NewIdentityPersister(db *pop.Connection) IdentityPersister {
	return &identityPersister{db: db}
}
