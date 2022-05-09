package handler

import (
	"github.com/gobuffalo/pop/v6"
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

	return h.persister.Transaction(func(tx *pop.Connection) error {
		p := h.persister.GetUserPersisterWithConnection(tx)
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
	})
}
