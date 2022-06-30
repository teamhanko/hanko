package handler

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"unicode/utf8"
)

type PasswordHandler struct {
	persister      persistence.Persister
	sessionManager session.Manager
	cfg            config.Password
}

func NewPasswordHandler(persister persistence.Persister, sessionManager session.Manager, cfg config.Password) *PasswordHandler {
	return &PasswordHandler{
		persister:      persister,
		sessionManager: sessionManager,
		cfg:            cfg,
	}
}

type PasswordSetBody struct {
	UserID   string `json:"user_id" validate:"required,uuid4"`
	Password string `json:"password" validate:"required"`
}

func (h *PasswordHandler) Set(c echo.Context) error {
	var body PasswordSetBody
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	sessionUserId, err := uuid.FromString(sessionToken.Subject())
	if err != nil {
		return dto.NewHTTPError(http.StatusBadRequest, "failed to parse userId as uuid").SetInternal(err)
	}

	pwBytes := []byte(body.Password)
	if utf8.RuneCountInString(body.Password) < h.cfg.MinPasswordLength { // use utf8.RuneCountInString, so utf8 characters would count as 1
		return dto.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("password must be at least %d characters long", h.cfg.MinPasswordLength))
	}
	if len(pwBytes) > 72 {
		return dto.NewHTTPError(http.StatusBadRequest, "password must not be longer than 72 bytes")
	}

	return h.persister.Transaction(func(tx *pop.Connection) error {
		user, err := h.persister.GetUserPersisterWithConnection(tx).Get(uuid.FromStringOrNil(body.UserID))
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if user == nil {
			return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(errors.New(fmt.Sprintf("user %s not found ", sessionUserId)))
		}

		if sessionUserId != user.ID {
			return dto.NewHTTPError(http.StatusForbidden).SetInternal(errors.New(fmt.Sprintf("session.userId %s tried to set password credentials for body.userId %s", sessionUserId, user.ID)))
		}

		pwPersister := h.persister.GetPasswordCredentialPersisterWithConnection(tx)
		pw, err := pwPersister.GetByUserID(user.ID)
		if err != nil {
			return fmt.Errorf("failed to get credential: %w", err)
		}

		hashedPassword, err := bcrypt.GenerateFromPassword(pwBytes, 12)
		if err != nil {
			return fmt.Errorf("failed to hash password: %s", err)
		}

		newPw := models.PasswordCredential{
			UserId:   uuid.FromStringOrNil(body.UserID),
			Password: string(hashedPassword),
		}

		if pw == nil {
			err = pwPersister.Create(newPw)
			if err != nil {
				return fmt.Errorf("failed to create password: %w", err)
			} else {
				return c.JSON(http.StatusCreated, nil)
			}
		} else {
			newPw.ID = pw.ID
			err = pwPersister.Update(newPw)
			if err != nil {
				return fmt.Errorf("failed to set password: %w", err)
			} else {
				return c.JSON(http.StatusOK, nil)
			}
		}
	})
}

type PasswordLoginBody struct {
	UserId   string `json:"user_id" validate:"required,uuid4"`
	Password string `json:"password" validate:"required"`
}

func (h *PasswordHandler) Login(c echo.Context) error {
	var body PasswordLoginBody
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	pwBytes := []byte(body.Password)
	if len(pwBytes) > 72 {
		return dto.NewHTTPError(http.StatusBadRequest, "password must not be longer than 72 bytes")
	}

	pw, err := h.persister.GetPasswordCredentialPersister().GetByUserID(uuid.FromStringOrNil(body.UserId))
	if pw == nil {
		return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(errors.New(fmt.Sprintf("no password credential found for: %s", body.UserId)))
	}

	if err != nil {
		return fmt.Errorf("error retrieving credential: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(pw.Password), pwBytes); err != nil {
		return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(err)
	}

	cookie, err := h.sessionManager.GenerateCookie(pw.UserId)
	if err != nil {
		return fmt.Errorf("failed to create session cookie: %w", err)
	}
	c.SetCookie(cookie)
	return c.String(http.StatusOK, "")
}
