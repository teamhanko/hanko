package handler

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/dto"
	"github.com/teamhanko/hanko/persistence"
	"github.com/teamhanko/hanko/persistence/models"
	"github.com/teamhanko/hanko/session"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type PasswordHandler struct {
	persister      persistence.Persister
	sessionManager session.Manager
}

func NewPasswordHandler(persister persistence.Persister, sessionManager session.Manager) *PasswordHandler {
	return &PasswordHandler{persister: persister, sessionManager: sessionManager}
}

type PasswordSetBody struct {
	UserID   string `json:"user_id" validate:"required,uuid4"`
	Password string `json:"password" validate:"required"`
}

func (h *PasswordHandler) Set(c echo.Context) error {
	var body PasswordSetBody
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := c.Validate(body); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	sessionUserId, err := uuid.FromString(sessionToken.Subject())
	if err != nil {
		return fmt.Errorf("failed to parse userId from JWT subject: %w", err)
	}

	return h.persister.Transaction(func(tx *pop.Connection) error {
		user, err := h.persister.GetUserPersisterWithConnection(tx).Get(uuid.FromStringOrNil(body.UserID))
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if user == nil {
			return c.JSON(http.StatusNotFound, dto.NewApiError(http.StatusNotFound))
		}

		if sessionUserId != user.ID {
			return c.JSON(http.StatusForbidden, dto.NewApiError(http.StatusForbidden))
		}

		pwPersister := h.persister.GetPasswordCredentialPersisterWithConnection(tx)
		pw, err := pwPersister.GetByUserID(user.ID)
		if err != nil {
			return fmt.Errorf("failed to get credential: %w", err)
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), 12)
		if err != nil {
			return errors.New("failed to create credential")
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
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := c.Validate(body); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	pw, err := h.persister.GetPasswordCredentialPersister().GetByUserID(uuid.FromStringOrNil(body.UserId))
	if pw == nil {
		return c.JSON(http.StatusNotFound, dto.NewApiError(http.StatusNotFound))
	}

	if err != nil {
		return fmt.Errorf("error retrieving credential: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(pw.Password), []byte(body.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, dto.NewApiError(http.StatusUnauthorized))
	}

	sessionToken, err := h.sessionManager.Generate(pw.UserId)
	if err != nil {
		return fmt.Errorf("failed to create session token: %w", err)
	}

	cookie := &http.Cookie{
		Name:     "hanko",
		Value:    sessionToken,
		Domain:   "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
	return c.String(http.StatusOK, "")
}
