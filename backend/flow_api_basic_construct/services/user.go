package services

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type User struct {
	persister persistence.Persister
}

func NewUserService(persister persistence.Persister) User {
	return User{
		persister,
	}
}

func (service User) CreateUser(id uuid.UUID, email string, emailVerified bool, username string, passkey *models.WebauthnCredential, password string) error {
	return service.persister.Transaction(func(tx *pop.Connection) error {
		// TODO: add audit log
		now := time.Now().UTC()
		err := service.persister.GetUserPersisterWithConnection(tx).Create(models.User{
			ID:        id,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return err
		}

		if email != "" {
			emailModel := models.NewEmail(&id, email)
			emailModel.Verified = emailVerified
			err = service.persister.GetEmailPersisterWithConnection(tx).Create(*emailModel)
			if err != nil {
				return err
			}

			primaryEmail := models.NewPrimaryEmail(emailModel.ID, id)
			err = service.persister.GetPrimaryEmailPersisterWithConnection(tx).Create(*primaryEmail)
			if err != nil {
				return err
			}
		}

		if username != "" {
			usernameModel := models.NewUsername(id, username)
			err = service.persister.GetUsernamePersisterWithConnection(tx).Create(*usernameModel)
			if err != nil {
				return err
			}
		}

		if passkey != nil {
			err = service.persister.GetWebauthnCredentialPersisterWithConnection(tx).Create(*passkey)
			if err != nil {
				return err
			}
		}

		if password != "" {
			err = service.persister.GetPasswordCredentialPersisterWithConnection(tx).Create(models.PasswordCredential{
				UserId:   id,
				Password: password,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
}
