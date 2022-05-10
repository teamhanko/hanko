package handler

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/dto"
	"github.com/teamhanko/hanko/persistence"
	"net/http"
)

type UserHandlerAdmin struct {
	persister persistence.Persister
}

func NewUserHandlerAdmin(persister persistence.Persister) *UserHandlerAdmin {
	return &UserHandlerAdmin{persister: persister}
}

func (h *UserHandlerAdmin) Delete(c echo.Context) error {
	userId, err := uuid.FromString(c.Param("id"))

	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest))
	}

	p := h.persister.GetUserPersister()
	user, err := p.Get(userId)
	if err != nil {
		return err
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, dto.NewApiError(http.StatusNotFound))
	}

	err = p.Delete(*user)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusNoContent, nil)
}

type UserPatchRequest struct {
	UserId   string `param:"id" validate:"required,uuid4"`
	Email    string `json:"email" validate:"omitempty,email"`
	Verified *bool  `json:"verified"`
}

func (h *UserHandlerAdmin) Patch(c echo.Context) error {
	var patchRequest UserPatchRequest
	if err := c.Bind(&patchRequest); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest))
	}

	if err := c.Validate(patchRequest); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	p := h.persister.GetUserPersister()
	user, err := p.Get(uuid.FromStringOrNil(patchRequest.UserId))
	if err != nil {
		return err
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, dto.NewApiError(http.StatusNotFound))
	}

	if patchRequest.Email != "" && patchRequest.Email != user.Email {
		maybeExistingUser, err := p.GetByEmail(patchRequest.Email)
		if err != nil {
			return err
		}

		if maybeExistingUser != nil {
			return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest).
				WithMessage("email address not available"))
		}

		user.Email = patchRequest.Email
	}

	if patchRequest.Verified != nil {
		user.Verified = *patchRequest.Verified
	}

	err = p.Update(*user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return c.JSON(http.StatusOK, nil)
}

type UserListRequest struct {
	PerPage int `query:"per_page"`
	Page    int `query:"page"`
}

func (h *UserHandlerAdmin) List(c echo.Context) error {
	var request UserListRequest
	err := echo.QueryParamsBinder(c).
		Int("per_page", &request.PerPage).
		Int("page", &request.Page).
		BindError()

	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest).WithMessage(err.Error()))
	}

	users, err := h.persister.GetUserPersister().List(request.Page, request.PerPage)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, users)
}
