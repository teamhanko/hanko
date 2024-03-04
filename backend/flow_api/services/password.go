package services

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrorPasswordInvalid = errors.New("password invalid")
)

type Password interface {
	VerifyPassword(userId uuid.UUID, password string) error
	RecoverPassword(userId uuid.UUID, newPassword string) error
	CreatePassword(userId uuid.UUID, newPassword string) error
	UpdatePassword(passwordCredentialModel *models.PasswordCredential, newPassword string) error
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

func (s password) VerifyPassword(userId uuid.UUID, password string) error {
	user, err := s.persister.GetUserPersister().Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return ErrorPasswordInvalid
	}

	pw, err := s.persister.GetPasswordCredentialPersister().GetByUserID(userId)
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

func (s password) RecoverPassword(userId uuid.UUID, newPassword string) error {
	passwordPersister := s.persister.GetPasswordCredentialPersister()

	passwordCredentialModel, err := passwordPersister.GetByUserID(userId)
	if err != nil {
		return fmt.Errorf("failed to get password credential by user id: %w", err)
	}

	if passwordCredentialModel == nil {
		err = s.CreatePassword(userId, newPassword)
	} else {
		err = s.UpdatePassword(passwordCredentialModel, newPassword)
	}

	if err != nil {
		return err
	}

	return nil
}

func (s password) CreatePassword(userId uuid.UUID, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return ErrorPasswordInvalid
	}

	passwordCredentialModel := models.NewPasswordCredential(userId, string(hashedPassword))

	err = s.persister.GetPasswordCredentialPersister().Create(*passwordCredentialModel)
	if err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	return nil
}

func (s password) UpdatePassword(passwordCredentialModel *models.PasswordCredential, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return ErrorPasswordInvalid
	}

	passwordCredentialModel.Password = string(hashedPassword)

	err = s.persister.GetPasswordCredentialPersister().Update(*passwordCredentialModel)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
