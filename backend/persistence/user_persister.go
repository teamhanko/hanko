package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type UserPersister interface {
	Get(uuid.UUID) (*models.User, error)
	GetByEmailAddress(string) (*models.User, error)
	GetByEmailAddressAndTenant(emailAddress string, tenantID *uuid.UUID) (*models.User, error)
	Create(models.User) error
	Update(models.User) error
	Delete(models.User) error
	List(page int, perPage int, userIDs []uuid.UUID, email string, username string, sortDirection string) ([]models.User, error)
	All() ([]models.User, error)
	Count(userIDs []uuid.UUID, email string, username string) (int, error)
	GetByUsername(username string) (*models.User, error)
	GetByUsernameAndTenant(username string, tenantID *uuid.UUID) (*models.User, error)
	// AdoptUserToTenant updates a user and all related records to belong to a tenant.
	// This is used when a global user (tenant_id = NULL) logs in with X-Tenant-ID.
	AdoptUserToTenant(userID uuid.UUID, tenantID uuid.UUID) error
}

type userPersister struct {
	db *pop.Connection
}

func NewUserPersister(db *pop.Connection) UserPersister {
	return &userPersister{db: db}
}

func (p *userPersister) Get(id uuid.UUID) (*models.User, error) {
	user := models.User{}

	eagerPreloadFields := []string{
		"Emails",
		"Emails.PrimaryEmail",
		"Emails.Identities.SamlIdentity",
		"WebauthnCredentials",
		"WebauthnCredentials.Transports",
		"Username",
		"PasswordCredential",
		"OTPSecret",
		"Metadata",
		"Identities",
	}

	err := p.db.EagerPreload(eagerPreloadFields...).Find(&user, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (p *userPersister) GetByEmailAddress(emailAddress string) (*models.User, error) {
	email := models.Email{}
	err := p.db.Eager().Where("address = (?)", emailAddress).First(&email)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email address: %w", err)
	}

	if email.UserID == nil {
		return nil, nil
	}

	return p.Get(*email.UserID)
}

func (p *userPersister) GetByEmailAddressAndTenant(emailAddress string, tenantID *uuid.UUID) (*models.User, error) {
	email := models.Email{}
	query := p.db.Eager().Where("address = (?)", emailAddress)
	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID.String())
	} else {
		query = query.Where("tenant_id IS NULL")
	}
	err := query.First(&email)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email address and tenant: %w", err)
	}

	if email.UserID == nil {
		return nil, nil
	}

	return p.Get(*email.UserID)
}

func (p *userPersister) GetByUsername(username string) (*models.User, error) {
	user := models.User{}
	err := p.db.EagerPreload(
		"Emails",
		"Emails.PrimaryEmail",
		"Emails.Identities",
		"WebauthnCredentials",
		"PasswordCredential",
		"Username",
		"OTPSecret",
		"Metadata").
		LeftJoin("usernames", "usernames.user_id = users.id").
		Where("usernames.username = (?)", username).
		First(&user)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (p *userPersister) GetByUsernameAndTenant(username string, tenantID *uuid.UUID) (*models.User, error) {
	user := models.User{}
	query := p.db.EagerPreload(
		"Emails",
		"Emails.PrimaryEmail",
		"Emails.Identities",
		"WebauthnCredentials",
		"PasswordCredential",
		"Username",
		"OTPSecret",
		"Metadata").
		LeftJoin("usernames", "usernames.user_id = users.id").
		Where("usernames.username = (?)", username)
	if tenantID != nil {
		query = query.Where("usernames.tenant_id = ?", tenantID.String())
	} else {
		query = query.Where("usernames.tenant_id IS NULL")
	}
	err := query.First(&user)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username and tenant: %w", err)
	}

	return &user, nil
}

func (p *userPersister) Create(user models.User) error {
	vErr, err := p.db.ValidateAndCreate(&user)
	if err != nil {
		return fmt.Errorf("failed to store user: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("user object validation failed: %w", vErr)
	}

	return nil
}

func (p *userPersister) Update(user models.User) error {
	vErr, err := p.db.ValidateAndUpdate(&user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("user object validation failed: %w", vErr)
	}

	return nil
}

// Delete deletes a user from the database including all information (e.g., emails, username, metadata)
// It must be called within a transaction otherwise some information might not be rolled back on an error.
func (p *userPersister) Delete(user models.User) error {
	primaryEmail := user.Emails.GetPrimary()

	if primaryEmail != nil {
		err := p.db.Destroy(primaryEmail.PrimaryEmail)
		if err != nil {
			return fmt.Errorf("failed to delete primary email: %w", err)
		}
	}

	err := p.db.Destroy(&user)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// AdoptUserToTenant updates a user and all related records to belong to a tenant.
// This migrates a global user (tenant_id = NULL) to a specific tenant.
func (p *userPersister) AdoptUserToTenant(userID uuid.UUID, tenantID uuid.UUID) error {
	// Update user
	err := p.db.RawQuery("UPDATE users SET tenant_id = ? WHERE id = ? AND tenant_id IS NULL", tenantID, userID).Exec()
	if err != nil {
		return fmt.Errorf("failed to update user tenant_id: %w", err)
	}

	// Update all related records
	tables := []string{"emails", "usernames", "identities", "webauthn_credentials",
		"otp_secrets", "password_credentials", "sessions"}
	for _, table := range tables {
		err := p.db.RawQuery(
			fmt.Sprintf("UPDATE %s SET tenant_id = ? WHERE user_id = ? AND tenant_id IS NULL", table),
			tenantID, userID,
		).Exec()
		if err != nil {
			return fmt.Errorf("failed to update %s tenant_id: %w", table, err)
		}
	}

	return nil
}

func (p *userPersister) List(page int, perPage int, userIDs []uuid.UUID, email string, username string, sortDirection string) ([]models.User, error) {
	users := []models.User{}

	query := p.db.
		Q().
		EagerPreload(
			"Emails",
			"Emails.PrimaryEmail",
			"WebauthnCredentials",
			"WebauthnCredentials.Transports",
			"Username").
		LeftJoin("emails", "emails.user_id = users.id").
		LeftJoin("usernames", "usernames.user_id = users.id")
	query = p.addQueryParamsToSqlQuery(query, userIDs, email, username)
	err := query.GroupBy("users.id").
		Having("count(emails.id) > 0 OR count(usernames.id) > 0").
		Order(fmt.Sprintf("users.created_at %s", sortDirection)).
		Paginate(page, perPage).
		All(&users)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return users, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	return users, nil
}

func (p *userPersister) All() ([]models.User, error) {
	users := []models.User{}

	err := p.db.EagerPreload(
		"Emails",
		"Emails.PrimaryEmail",
		"Emails.Identities",
		"WebauthnCredentials",
		"WebauthnCredentials.Transports",
		"Username",
	).All(&users)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return users, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	return users, nil
}

func (p *userPersister) Count(userIDs []uuid.UUID, email string, username string) (int, error) {
	query := p.db.
		Q().
		LeftJoin("emails", "emails.user_id = users.id").
		LeftJoin("usernames", "usernames.user_id = users.id")
	query = p.addQueryParamsToSqlQuery(query, userIDs, email, username)
	count, err := query.GroupBy("users.id").
		Having("count(emails.id) > 0 OR count(usernames.id) > 0").
		Count(&models.User{})
	if err != nil {
		return 0, fmt.Errorf("failed to get user count: %w", err)
	}

	return count, nil
}

func (p *userPersister) addQueryParamsToSqlQuery(query *pop.Query, userIDs []uuid.UUID, email string, username string) *pop.Query {
	if email != "" && username != "" {
		query = query.Where("emails.address LIKE ? OR usernames.username LIKE ?", "%"+email+"%", "%"+username+"%")
	} else if email != "" {
		query = query.Where("emails.address LIKE ?", "%"+email+"%")
	} else if username != "" {
		query = query.Where("usernames.username LIKE ?", "%"+username+"%")
	}

	if len(userIDs) > 0 {
		query = query.Where("users.id in (?)", userIDs)
	}

	return query
}
