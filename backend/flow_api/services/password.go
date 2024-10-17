package services

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrorPasswordInvalid = errors.New("password invalid")
)

type Password interface {
	VerifyPassword(tx *pop.Connection, userId uuid.UUID, password string) error
	RecoverPassword(tx *pop.Connection, userId uuid.UUID, newPassword string) error
	CreatePassword(tx *pop.Connection, userId uuid.UUID, newPassword string) error
	UpdatePassword(tx *pop.Connection, passwordCredentialModel *models.PasswordCredential, newPassword string) error
}

type password struct {
	persister persistence.Persister
	cfg       config.Config
}

func NewPasswordService(cfg config.Config, persister persistence.Persister) Password {
	return &password{
		persister,
		cfg,
	}
}

func (s password) VerifyPassword(tx *pop.Connection, userId uuid.UUID, password string) error {
	user, err := s.persister.GetUserPersisterWithConnection(tx).Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return ErrorPasswordInvalid
	}

	pw, err := s.persister.GetPasswordCredentialPersisterWithConnection(tx).GetByUserID(userId)
	if err != nil {
		return fmt.Errorf("error retrieving password credential: %w", err)
	}

	if pw == nil {
		return ErrorPasswordInvalid
	}

	if err = bcrypt.CompareHashAndPassword([]byte(pw.Password), []byte(password)); err != nil {
		return ErrorPasswordInvalid
	}

	return nil
}

func (s password) RecoverPassword(tx *pop.Connection, userId uuid.UUID, newPassword string) error {
	passwordPersister := s.persister.GetPasswordCredentialPersisterWithConnection(tx)

	passwordCredentialModel, err := passwordPersister.GetByUserID(userId)
	if err != nil {
		return fmt.Errorf("failed to get password credential by user id: %w", err)
	}

	if passwordCredentialModel == nil {
		err = s.CreatePassword(tx, userId, newPassword)
	} else {
		err = s.UpdatePassword(tx, passwordCredentialModel, newPassword)
	}

	if err != nil {
		return err
	}

	return nil
}

func (s password) CreatePassword(tx *pop.Connection, userId uuid.UUID, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return ErrorPasswordInvalid
	}

	passwordCredentialModel := models.NewPasswordCredential(userId, string(hashedPassword))

	err = s.persister.GetPasswordCredentialPersisterWithConnection(tx).Create(*passwordCredentialModel)
	if err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	return nil
}

func (s password) UpdatePassword(tx *pop.Connection, passwordCredentialModel *models.PasswordCredential, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return ErrorPasswordInvalid
	}

	passwordCredentialModel.Password = string(hashedPassword)
	passwordCredentialModel.UpdatedAt = time.Now().UTC()

	err = s.persister.GetPasswordCredentialPersisterWithConnection(tx).Update(*passwordCredentialModel)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
