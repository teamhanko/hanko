package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgconn"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/teamhanko/hanko/backend/crypto"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/dto/admin"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"golang.org/x/crypto/bcrypt"
)

type PasslinkHandlerAdmin struct {
	passlinkGenerator crypto.PasslinkGenerator
	persister         persistence.Persister
}

func NewPasslinkHandlerAdmin(persister persistence.Persister) *PasslinkHandlerAdmin {
	return &PasslinkHandlerAdmin{persister: persister}
}

func (h *PasslinkHandlerAdmin) Delete(c echo.Context) error {
	passlinkId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse passlinkId as uuid").SetInternal(err)
	}

	p := h.persister.GetPasslinkPersister()
	passlink, err := p.Get(passlinkId)
	if err != nil {
		return fmt.Errorf("failed to get passlink: %w", err)
	}

	if passlink == nil {
		return echo.NewHTTPError(http.StatusNotFound, "passlink not found")
	}

	err = p.Delete(*passlink)
	if err != nil {
		return fmt.Errorf("failed to delete passlink: %w", err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *PasslinkHandlerAdmin) Get(c echo.Context) error {
	passlinkId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse passlinkId as uuid").SetInternal(err)
	}

	p := h.persister.GetPasslinkPersister()
	passlink, err := p.Get(passlinkId)
	if err != nil {
		return fmt.Errorf("failed to get passlink: %w", err)
	}

	if passlink == nil {
		return echo.NewHTTPError(http.StatusNotFound, "passlink not found")
	}

	return c.JSON(http.StatusOK, admin.FromPasslinkModel(*passlink))
}

func (h *PasslinkHandlerAdmin) Create(c echo.Context) error {
	var body admin.CreatePasslink
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	// if no passlinkID is provided, create a new one
	if body.ID == nil || body.ID.IsNil() {
		passlinkId, err := uuid.NewV4()
		if err != nil {
			return fmt.Errorf("failed to create new passlinkId: %w", err)
		}
		body.ID = &passlinkId
	}

	now := time.Now().UTC()
	token, err := h.passlinkGenerator.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate passlink: %w", err)
	}
	tokenHashed, err := bcrypt.GenerateFromPassword([]byte(token), 12)
	if err != nil {
		return fmt.Errorf("failed to hash passlink: %w", err)
	}

	err = h.persister.GetConnection().Transaction(func(tx *pop.Connection) error {
		passlink := models.Passlink{
			ID:         *body.ID,
			UserID:     body.UserID,  // FIXME: validate us
			EmailID:    body.EmailID, // FIXME: validate emailID
			IP:         c.RealIP(),
			TTL:        body.TTL,
			LoginCount: 0,
			Reusable:   body.Reusable,
			Token:      string(tokenHashed),
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		err := tx.Create(&passlink)
		if err != nil {
			var pgErr *pgconn.PgError
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" {
					return echo.NewHTTPError(http.StatusConflict, fmt.Errorf("failed to create passlink with id '%v': %w", passlink.ID, fmt.Errorf("passlink already exists")))
				}
			} else if errors.As(err, &mysqlErr) {
				if mysqlErr.Number == 1062 {
					return echo.NewHTTPError(http.StatusConflict, fmt.Errorf("failed to create passlink with id '%v': %w", passlink.ID, fmt.Errorf("passlink already exists")))
				}
			}
			return fmt.Errorf("failed to create passlink with id '%v': %w", passlink.ID, err)
		}

		return nil
	})

	if httpError, ok := err.(*echo.HTTPError); ok {
		return httpError
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	p := h.persister.GetPasslinkPersister()
	passlink, err := p.Get(*body.ID)
	if err != nil {
		return fmt.Errorf("failed to get passlink: %w", err)
	}

	if passlink == nil {
		return echo.NewHTTPError(http.StatusNotFound, "passlink not found")
	}

	passlinkDto := admin.FromPasslinkModel(*passlink)

	return c.JSON(http.StatusOK, passlinkDto)
}
