package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrorPasswordInvalid = errors.New("password invalid")
)

type Password interface {
	VerifyPassword(tx *pop.Connection, userId uuid.UUID, password string, tenantID *uuid.UUID) error
	RecoverPassword(tx *pop.Connection, userId uuid.UUID, newPassword string, tenantID *uuid.UUID) error
	CreatePassword(tx *pop.Connection, userId uuid.UUID, newPassword string, tenantID *uuid.UUID) error
	UpdatePassword(tx *pop.Connection, passwordCredentialModel *models.PasswordCredential, newPassword string) error
}

type password struct {
	persister persistence.Persister
}

func NewPasswordService(persister persistence.Persister) Password {
	return &password{
		persister,
	}
}

func (s password) VerifyPassword(tx *pop.Connection, userId uuid.UUID, password string, tenantID *uuid.UUID) error {
	user, err := s.persister.GetUserPersisterWithConnection(tx).Get(userId, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return ErrorPasswordInvalid
	}

	pw, err := s.persister.GetPasswordCredentialPersisterWithConnection(tx).GetByUserID(userId, tenantID)
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

func (s password) RecoverPassword(tx *pop.Connection, userId uuid.UUID, newPassword string, tenantID *uuid.UUID) error {
	passwordPersister := s.persister.GetPasswordCredentialPersisterWithConnection(tx)

	passwordCredentialModel, err := passwordPersister.GetByUserID(userId, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get password credential by user id: %w", err)
	}

	if passwordCredentialModel == nil {
		err = s.CreatePassword(tx, userId, newPassword, tenantID)
	} else {
		err = s.UpdatePassword(tx, passwordCredentialModel, newPassword)
	}

	if err != nil {
		return err
	}

	return nil
}

func (s password) CreatePassword(tx *pop.Connection, userId uuid.UUID, newPassword string, tenantID *uuid.UUID) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return ErrorPasswordInvalid
	}

	passwordCredentialModel := models.NewPasswordCredential(userId, string(hashedPassword), tenantID)

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
