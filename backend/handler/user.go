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
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
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
			return fmt.Errorf("failed to get user: %w", err)
		}

		if email != nil {
			if email.UserID != nil {
				// The email already exists and is assigned already.
				return dto.NewHTTPError(http.StatusConflict).SetInternal(errors.New(fmt.Sprintf("user with email %s already exists", body.Email)))
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

			c.SetCookie(cookie)

			if h.cfg.Session.EnableAuthTokenHeader {
				c.Response().Header().Set("X-Auth-Token", token)
				c.Response().Header().Set("Access-Control-Expose-Headers", "X-Auth-Token")
			}
		}

		err = h.auditLogger.Create(c, models.AuditLogUserCreated, &newUser, nil)
		if err != nil {
			return fmt.Errorf("failed to write audit log: %w", err)
		}

		// this cookie is a workaround for older hanko element versions,
		// because else the backend would not know where to send the first passcode
		c.SetCookie(&http.Cookie{
			Name:     "hanko_email_id",
			Value:    email.ID.String(),
			Domain:   h.cfg.Session.Cookie.Domain,
			Secure:   h.cfg.Session.Cookie.Secure,
			HttpOnly: h.cfg.Session.Cookie.HttpOnly,
		})

		return c.JSON(http.StatusOK, dto.CreateUserResponse{
			ID:      newUser.ID,
			UserID:  newUser.ID,
			EmailID: email.ID,
		})
	})
}

func (h *UserHandler) Get(c echo.Context) error {
	userId := c.Param("id")

	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	if sessionToken.Subject() != userId {
		return dto.NewHTTPError(http.StatusForbidden).SetInternal(errors.New(fmt.Sprintf("user %s tried to get user %s", sessionToken.Subject(), userId)))
	}

	user, err := h.persister.GetUserPersister().Get(uuid.FromStringOrNil(userId))
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return dto.NewHTTPError(http.StatusNotFound).SetInternal(errors.New("user not found"))
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
		return dto.NewHTTPError(http.StatusNotFound).SetInternal(errors.New("user not found"))
	}

	credentials, err := h.persister.GetWebauthnCredentialPersister().GetFromUser(*email.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
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
