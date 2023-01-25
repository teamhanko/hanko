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
	Create(models.User) error
	Update(models.User) error
	Delete(models.User) error
	List(page int, perPage int) ([]models.User, error)
	Count() (int, error)
}

type userPersister struct {
	db *pop.Connection
}

func NewUserPersister(db *pop.Connection) UserPersister {
	return &userPersister{db: db}
}

func (p *userPersister) Get(id uuid.UUID) (*models.User, error) {
	user := models.User{}
	err := p.db.EagerPreload("Emails", "Emails.PrimaryEmail", "WebauthnCredentials").Find(&user, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (p *userPersister) GetByEmail(email string) (*models.User, error) {
	user := models.User{}
	err := p.db.Eager().Where("email = (?)", email).First(&user)
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

func (p *userPersister) List(page int, perPage int) ([]models.User, error) {
	users := []models.User{}

	err := p.db.Q().Paginate(page, perPage).All(&users)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return users, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	return users, nil
}

func (p *userPersister) Count() (int, error) {
	count, err := p.db.Count(&models.User{})
	if err != nil {
		return 0, fmt.Errorf("failed to get user count: %w", err)
	}

	return count, nil
}
