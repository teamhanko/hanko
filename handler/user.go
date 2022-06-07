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
	"net/http"
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
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := c.Validate(body); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	return h.persister.Transaction(func(tx *pop.Connection) error {
		user, err := h.persister.GetUserPersisterWithConnection(tx).GetByEmail(body.Email)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if user != nil {
			return c.JSON(http.StatusConflict, dto.NewApiError(http.StatusConflict))
		}

		newUser := models.NewUser(body.Email)
		err = h.persister.GetUserPersisterWithConnection(tx).Create(newUser)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, newUser)
	})
}

func (h *UserHandler) Get(c echo.Context) error {
	userId := c.Param("id")

	user, err := h.persister.GetUserPersister().Get(uuid.FromStringOrNil(userId))
	if err != nil {
		return err
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, dto.NewApiError(http.StatusNotFound))
	}

	return c.JSON(http.StatusOK, user)
}

type UserGetByEmailBody struct {
	Email string `json:"email" validate:"required,email"`
}

func (h *UserHandler) GetUserIdByEmail(c echo.Context) error {
	var request UserGetByEmailBody
	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := c.Validate(request); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	user, err := h.persister.GetUserPersister().GetByEmail(request.Email)
	if err != nil {
		return err
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, dto.NewApiError(http.StatusNotFound))
	}

	return c.JSON(http.StatusOK, struct {
		UserId   string `json:"id"`
		Verified bool   `json:"verified"`
	}{
		UserId:   user.ID.String(),
		Verified: user.Verified,
	})
}

func (h *UserHandler) Me(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}

	return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/users/%s", sessionToken.Subject()))
}
