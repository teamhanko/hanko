package hooks

import (
	"encoding/json"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/dto/intern"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"time"
)

func NewBeforeSuccess(persister persistence.Persister, sessionManager session.Manager, httpContext echo.Context) BeforeSuccess {
	return BeforeSuccess{
		persister,
		sessionManager,
		httpContext,
	}
}

type BeforeSuccess struct {
	persister      persistence.Persister
	sessionManager session.Manager
	httpContext    echo.Context
}

func (m BeforeSuccess) Execute(c flowpilot.HookExecutionContext) error {
	userId, err := uuid.NewV4()
	if err != nil {
		return err
	}
	if c.Stash().Get("user_id").Exists() {
		userId, err = uuid.FromString(c.Stash().Get("user_id").String())
		if err != nil {
			return err
		}
	}

	passkeyCredentialStr := c.Stash().Get("passkey_credential").String()
	var passkeyCredential webauthnLib.Credential
	err = json.Unmarshal([]byte(passkeyCredentialStr), &passkeyCredential)
	if err != nil {
		return err
	}
	passkeyBackupEligible := c.Stash().Get("passkey_backup_eligible").Bool()
	passkeyBackupState := c.Stash().Get("passkey_backup_state").Bool()

	credentialModel := intern.WebauthnCredentialToModel(&passkeyCredential, userId, passkeyBackupEligible, passkeyBackupState)
	err = m.CreateUser(
		userId,
		c.Stash().Get("email").String(),
		c.Stash().Get("email_verified").Bool(),
		c.Stash().Get("username").String(),
		credentialModel,
		c.Stash().Get("new_password").String(),
	)
	if err != nil {
		return err
	}

	sessionToken, err := m.sessionManager.GenerateJWT(userId)
	if err != nil {
		return err
	}
	cookie, err := m.sessionManager.GenerateCookie(sessionToken)
	if err != nil {
		return err
	}

	m.httpContext.SetCookie(cookie)

	return nil
}

func (m BeforeSuccess) CreateUser(id uuid.UUID, email string, emailVerified bool, username string, passkey *models.WebauthnCredential, password string) error {
	return m.persister.Transaction(func(tx *pop.Connection) error {
		// TODO: add audit log
		now := time.Now().UTC()
		err := m.persister.GetUserPersisterWithConnection(tx).Create(models.User{
			ID:        id,
			Username:  username,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return err
		}

		if email != "" {
			emailModel := models.NewEmail(&id, email)
			emailModel.Verified = emailVerified
			err = m.persister.GetEmailPersisterWithConnection(tx).Create(*emailModel)
			if err != nil {
				return err
			}

			primaryEmail := models.NewPrimaryEmail(emailModel.ID, id)
			err = m.persister.GetPrimaryEmailPersisterWithConnection(tx).Create(*primaryEmail)
			if err != nil {
				return err
			}
		}

		if passkey != nil {
			err = m.persister.GetWebauthnCredentialPersisterWithConnection(tx).Create(*passkey)
			if err != nil {
				return err
			}
		}

		if password != "" {
			err = m.persister.GetPasswordCredentialPersisterWithConnection(tx).Create(models.PasswordCredential{
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
