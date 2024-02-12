package handler

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/dto/admin"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"github.com/teamhanko/hanko/backend/webhooks/utils"
	"net/http"
	"strings"
)

type UserHandler struct {
	persister      persistence.Persister
	sessionManager session.Manager
	auditLogger    auditlog.Logger
	cfg            *config.Config
}

func NewUserHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, auditLogger auditlog.Logger) *UserHandler {
	return &UserHandler{
		persister:      persister,
		auditLogger:    auditLogger,
		sessionManager: sessionManager,
		cfg:            cfg,
	}
}

type UserCreateBody struct {
	Email string `json:"email" validate:"required,email"`
}

func (h *UserHandler) Create(c echo.Context) error {
	if !h.cfg.Account.AllowSignup {
		return echo.NewHTTPError(http.StatusForbidden).SetInternal(errors.New("account signup is disabled"))
	}

	var body UserCreateBody
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	body.Email = strings.ToLower(body.Email)

	return h.persister.Transaction(func(tx *pop.Connection) error {
		newUser := models.NewUser()
		err := h.persister.GetUserPersisterWithConnection(tx).Create(newUser)
		if err != nil {
			return fmt.Errorf("failed to store user: %w", err)
		}

		email, err := h.persister.GetEmailPersisterWithConnection(tx).FindByAddress(body.Email)
		if err != nil {
			return fmt.Errorf("failed to get email: %w", err)
		}

		if email != nil {
			if email.UserID != nil {
				// The email already exists and is assigned already.
				return echo.NewHTTPError(http.StatusConflict).SetInternal(errors.New(fmt.Sprintf("user with email %s already exists", body.Email)))
			}

			if !h.cfg.Emails.RequireVerification {
				// Assign the email address to the user because it's currently unassigned and email verification is turned off.
				email.UserID = &newUser.ID
				err = h.persister.GetEmailPersisterWithConnection(tx).Update(*email)
				if err != nil {
					return fmt.Errorf("failed to update email address: %w", err)
				}
			}
		} else {
			// The email address does not exist, create a new one.
			if h.cfg.Emails.RequireVerification {
				// The email can only be assigned to the user via passcode verification.
				email = models.NewEmail(nil, body.Email)
			} else {
				email = models.NewEmail(&newUser.ID, body.Email)
			}

			err = h.persister.GetEmailPersisterWithConnection(tx).Create(*email)
			if err != nil {
				return fmt.Errorf("failed to store user: %w", err)
			}
		}

		if !h.cfg.Emails.RequireVerification {
			primaryEmail := models.NewPrimaryEmail(email.ID, newUser.ID)
			err = h.persister.GetPrimaryEmailPersisterWithConnection(tx).Create(*primaryEmail)
			if err != nil {
				return fmt.Errorf("failed to store primary email: %w", err)
			}

			token, err := h.sessionManager.GenerateJWT(newUser.ID)
			if err != nil {
				return fmt.Errorf("failed to generate jwt: %w", err)
			}

			cookie, err := h.sessionManager.GenerateCookie(token)
			if err != nil {
				return fmt.Errorf("failed to create session token: %w", err)
			}

			c.Response().Header().Set("X-Session-Lifetime", fmt.Sprintf("%d", cookie.MaxAge))

			if h.cfg.Session.EnableAuthTokenHeader {
				c.Response().Header().Set("X-Auth-Token", token)
			} else {
				c.SetCookie(cookie)
			}
		}

		err = h.auditLogger.CreateWithConnection(tx, c, models.AuditLogUserCreated, &newUser, nil)
		if err != nil {
			return fmt.Errorf("failed to write audit log: %w", err)
		}

		// This cookie is a workaround for hanko element versions before 0.1.0-alpha,
		// because else the backend would not know where to send the first passcode.
		c.SetCookie(&http.Cookie{
			Name:     "hanko_email_id",
			Value:    email.ID.String(),
			Domain:   h.cfg.Session.Cookie.Domain,
			Secure:   h.cfg.Session.Cookie.Secure,
			HttpOnly: h.cfg.Session.Cookie.HttpOnly,
			SameSite: http.SameSiteNoneMode,
		})

		newUserDto := dto.CreateUserResponse{
			ID:      newUser.ID,
			UserID:  newUser.ID,
			EmailID: email.ID,
		}

		if !h.cfg.Emails.RequireVerification {
			err = utils.TriggerWebhooks(c, events.UserCreate, admin.FromUserModel(newUser))
			if err != nil {
				c.Logger().Warn(err)
			}
		}

		return c.JSON(http.StatusOK, newUserDto)
	})
}

func (h *UserHandler) Get(c echo.Context) error {
	userId := c.Param("id")

	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	if sessionToken.Subject() != userId {
		return echo.NewHTTPError(http.StatusForbidden).SetInternal(errors.New(fmt.Sprintf("user %s tried to get user %s", sessionToken.Subject(), userId)))
	}

	user, err := h.persister.GetUserPersister().Get(uuid.FromStringOrNil(userId))
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound).SetInternal(errors.New("user not found"))
	}

	var emailAddress *string
	if e := user.Emails.GetPrimary(); e != nil {
		emailAddress = &e.Address
	}

	return c.JSON(http.StatusOK, dto.GetUserResponse{
		ID:                  user.ID,
		WebauthnCredentials: user.WebauthnCredentials,
		Email:               emailAddress,
		CreatedAt:           user.CreatedAt,
		UpdatedAt:           user.UpdatedAt,
	})
}

type UserGetByEmailBody struct {
	Email string `json:"email" validate:"required,email"`
}

func (h *UserHandler) GetUserIdByEmail(c echo.Context) error {
	var request UserGetByEmailBody
	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(request); err != nil {
		return dto.ToHttpError(err)
	}

	emailAddress := strings.ToLower(request.Email)
	email, err := h.persister.GetEmailPersister().FindByAddress(emailAddress)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if email == nil || email.UserID == nil {
		return echo.NewHTTPError(http.StatusNotFound).SetInternal(errors.New("user not found"))
	}

	credentials, err := h.persister.GetWebauthnCredentialPersister().GetFromUser(*email.UserID)
	if err != nil {
		return fmt.Errorf("failed to get webauthn credentials: %w", err)
	}

	return c.JSON(http.StatusOK, dto.UserInfoResponse{
		ID:                    *email.UserID,
		Verified:              email.Verified,
		EmailID:               email.ID,
		HasWebauthnCredential: len(credentials) > 0,
	})
}

func (h *UserHandler) Me(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}

	return c.JSON(http.StatusOK, map[string]string{"id": sessionToken.Subject()})
}

func (h *UserHandler) Delete(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	userId, err := uuid.FromString(sessionToken.Subject())
	if err != nil {
		return fmt.Errorf("failed to parse subject as uuid: %w", err)
	}

	return h.persister.Transaction(func(tx *pop.Connection) error {
		user, err := h.persister.GetUserPersisterWithConnection(tx).Get(userId)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if user == nil {
			return fmt.Errorf("unknown user")
		}

		err = h.persister.GetUserPersisterWithConnection(tx).Delete(*user)
		if err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		err = h.auditLogger.CreateWithConnection(tx, c, models.AuditLogUserDeleted, user, nil)
		if err != nil {
			return fmt.Errorf("failed to write audit log: %w", err)
		}

		cookie, err := h.sessionManager.DeleteCookie()
		if err != nil {
			return fmt.Errorf("failed to create session token: %w", err)
		}

		c.SetCookie(cookie)

		err = utils.TriggerWebhooks(c, events.UserDelete, admin.FromUserModel(*user))
		if err != nil {
			c.Logger().Warn(err)
		}

		return c.NoContent(http.StatusNoContent)
	})
}

func (h *UserHandler) Logout(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	userId := uuid.FromStringOrNil(sessionToken.Subject())

	user, err := h.persister.GetUserPersister().Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	err = h.auditLogger.Create(c, models.AuditLogUserLoggedOut, user, nil)
	if err != nil {
		return fmt.Errorf("failed to write audit log: %w", err)
	}

	cookie, err := h.sessionManager.DeleteCookie()
	if err != nil {
		return fmt.Errorf("failed to create session token: %w", err)
	}

	c.SetCookie(cookie)

	return c.NoContent(http.StatusNoContent)
}
