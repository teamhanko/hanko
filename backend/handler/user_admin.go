package handler

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/pagination"
	"github.com/teamhanko/hanko/backend/persistence"
	"net/http"
	"net/url"
	"strconv"
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
		return dto.NewHTTPError(http.StatusBadRequest, "failed to parse userId as uuid").SetInternal(err)
	}

	p := h.persister.GetUserPersister()
	user, err := p.Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return dto.NewHTTPError(http.StatusNotFound, "user not found")
	}

	err = p.Delete(*user)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return c.JSON(http.StatusNoContent, nil)
}

type UserPatchRequest struct {
	UserId   string `param:"id" validate:"required,uuid4"`
	Email    string `json:"email" validate:"omitempty,email"`
	Verified *bool  `json:"verified"`
}

type UserListRequest struct {
	PerPage int `query:"per_page"`
	Page    int `query:"page"`
}

func (h *UserHandlerAdmin) List(c echo.Context) error {
	var request UserListRequest
	err := (&echo.DefaultBinder{}).BindQueryParams(c, &request)
	if err != nil {
		return dto.ToHttpError(err)
	}

	if request.Page == 0 {
		request.Page = 1
	}

	if request.PerPage == 0 {
		request.PerPage = 20
	}

	users, err := h.persister.GetUserPersister().List(request.Page, request.PerPage)
	if err != nil {
		return fmt.Errorf("failed to get list of users: %w", err)
	}

	userCount, err := h.persister.GetUserPersister().Count()
	if err != nil {
		return fmt.Errorf("failed to get total count of users: %w", err)
	}

	u, _ := url.Parse(fmt.Sprintf("%s://%s%s", c.Scheme(), c.Request().Host, c.Request().RequestURI))

	c.Response().Header().Set("Link", pagination.CreateHeader(u, userCount, request.Page, request.PerPage))
	c.Response().Header().Set("X-Total-Count", strconv.FormatInt(int64(userCount), 10))

	return c.JSON(http.StatusOK, users)
}
