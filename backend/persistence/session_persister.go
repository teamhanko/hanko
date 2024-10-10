package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type SessionPersister interface {
	Create(session models.Session) error
	Get(id uuid.UUID) (*models.Session, error)
	Update(session models.Session) error
	List(userID uuid.UUID) ([]models.Session, error)
	ListActive(userID uuid.UUID) ([]models.Session, error)
	Delete(session models.Session) error
}

type sessionPersister struct {
	db *pop.Connection
}

func NewSessionPersister(db *pop.Connection) SessionPersister {
	return &sessionPersister{db: db}
}

func (p *sessionPersister) Create(session models.Session) error {
	vErr, err := p.db.ValidateAndCreate(&session)
	if err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}
	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("session object validation failed: %w", vErr)
	}

	return nil
}

func (p *sessionPersister) Get(id uuid.UUID) (*models.Session, error) {
	session := models.Session{}
	err := p.db.Eager().Find(&session, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

func (p *sessionPersister) Update(session models.Session) error {
	vErr, err := p.db.ValidateAndUpdate(&session)
	if err != nil {
		return err
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("session object validation failed: %w", vErr)
	}

	return nil
}

func (p *sessionPersister) List(userID uuid.UUID) ([]models.Session, error) {
	sessions := []models.Session{}

	err := p.db.Q().Where("user_id = ?", userID).All(&sessions)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return sessions, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sessions: %w", err)
	}

	return sessions, nil
}

func (p *sessionPersister) ListActive(userID uuid.UUID) ([]models.Session, error) {
	sessions := []models.Session{}

	err := p.db.Q().Where("user_id = ?", userID).Where("expires_at > ?", time.Now()).Order("created_at desc").All(&sessions)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return sessions, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sessions: %w", err)
	}

	return sessions, nil
}

func (p *sessionPersister) Delete(session models.Session) error {
	err := p.db.Eager().Destroy(&session)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}
