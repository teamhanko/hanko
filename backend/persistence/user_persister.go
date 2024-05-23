package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type UserPersister interface {
	Get(uuid.UUID) (*models.User, error)
	GetByEmailAddress(string) (*models.User, error)
	Create(models.User) error
	Update(models.User) error
	Delete(models.User) error
	List(page int, perPage int, userId uuid.UUID, email string, sortDirection string) ([]models.User, error)
	All() ([]models.User, error)
	Count(userId uuid.UUID, email string) (int, error)
	GetByUsername(username string) (*models.User, error)
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
		"Emails.Identities",
		"WebauthnCredentials",
		"WebauthnCredentials.Transports",
		"PasswordCredential"}

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
	err := p.db.Where("address = (?)", emailAddress).First(&email)

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

func (p *userPersister) GetByUsername(username string) (*models.User, error) {
	user := models.User{}
	err := p.db.EagerPreload("Emails", "Emails.PrimaryEmail", "Emails.Identities", "WebauthnCredentials", "PasswordCredential").Where("username = (?)", username).First(&user)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
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

func (p *userPersister) Delete(user models.User) error {
	err := p.db.Destroy(&user)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (p *userPersister) List(page int, perPage int, userId uuid.UUID, email string, sortDirection string) ([]models.User, error) {
	users := []models.User{}

	query := p.db.
		Q().
		EagerPreload("Emails", "Emails.PrimaryEmail", "WebauthnCredentials").
		LeftJoin("emails", "emails.user_id = users.id")
	query = p.addQueryParamsToSqlQuery(query, userId, email)
	err := query.GroupBy("users.id").
		Having("count(emails.id) > 0").
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
	err := p.db.EagerPreload("Emails", "Emails.PrimaryEmail", "Emails.Identities", "WebauthnCredentials").All(&users)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return users, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	return users, nil
}

func (p *userPersister) Count(userId uuid.UUID, email string) (int, error) {
	query := p.db.
		Q().
		LeftJoin("emails", "emails.user_id = users.id")
	query = p.addQueryParamsToSqlQuery(query, userId, email)
	count, err := query.GroupBy("users.id").
		Having("count(emails.id) > 0").
		Count(&models.User{})
	if err != nil {
		return 0, fmt.Errorf("failed to get user count: %w", err)
	}

	return count, nil
}

func (p *userPersister) addQueryParamsToSqlQuery(query *pop.Query, userId uuid.UUID, email string) *pop.Query {
	if email != "" {
		query = query.Where("emails.address LIKE ?", "%"+email+"%")
	}
	if !userId.IsNil() {
		query = query.Where("users.id = ?", userId)
	}

	return query
}
