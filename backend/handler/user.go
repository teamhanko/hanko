package handler

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"net/http"
	"strings"
)

type UserHandler struct {
	persister persistence.Persister
}

func NewUserHandler(persister persistence.Persister) *UserHandler {
	return &UserHandler{persister: persister}
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
		user, err := h.persister.GetUserPersisterWithConnection(tx).GetByEmail(body.Email)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if user != nil {
			return dto.NewHTTPError(http.StatusConflict).SetInternal(errors.New(fmt.Sprintf("user with email %s already exists", user.Email)))
		}

		newUser := models.NewUser(body.Email)
		err = h.persister.GetUserPersisterWithConnection(tx).Create(newUser)
		if err != nil {
			return fmt.Errorf("failed to store user: %w", err)
		}

		return c.JSON(http.StatusOK, newUser)
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

	return c.JSON(http.StatusOK, user)
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

	user, err := h.persister.GetUserPersister().GetByEmail(strings.ToLower(request.Email))
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return dto.NewHTTPError(http.StatusNotFound).SetInternal(errors.New("user not found"))
	}

	return c.JSON(http.StatusOK, struct {
		UserId                string `json:"id"`
		Verified              bool   `json:"verified"`
		HasWebauthnCredential bool   `json:"has_webauthn_credential"`
	}{
		UserId:                user.ID.String(),
		Verified:              user.Verified,
		HasWebauthnCredential: len(user.WebauthnCredentials) > 0,
	})
}

func (h *UserHandler) Me(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}

	return c.JSON(http.StatusOK, map[string]string{"id": sessionToken.Subject()})
}
