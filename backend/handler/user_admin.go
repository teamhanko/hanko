package handler

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgconn"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/dto/admin"
	"github.com/teamhanko/hanko/backend/pagination"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"github.com/teamhanko/hanko/backend/webhooks/utils"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse userId as uuid").SetInternal(err)
	}

	p := h.persister.GetUserPersister()
	user, err := p.Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	err = p.Delete(*user)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	err = utils.TriggerWebhooks(c, events.UserDelete, admin.FromUserModel(*user))
	if err != nil {
		c.Logger().Warn(err)
	}

	return c.NoContent(http.StatusNoContent)
}

type UserListRequest struct {
	PerPage       int    `query:"per_page"`
	Page          int    `query:"page"`
	Email         string `query:"email"`
	UserId        string `query:"user_id"`
	Username      string `query:"username"`
	SortDirection string `query:"sort_direction"`
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

	userId := uuid.Nil
	if request.UserId != "" {
		userId, err = uuid.FromString(request.UserId)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "failed to parse user_id as uuid").SetInternal(err)
		}
	}

	if request.SortDirection == "" {
		request.SortDirection = "desc"
	}

	switch request.SortDirection {
	case "desc", "asc":
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "sort_direction must be desc or asc")
	}

	email := strings.ToLower(request.Email)
	username := strings.ToLower(request.Username)

	users, err := h.persister.GetUserPersister().List(request.Page, request.PerPage, userId, email, username, request.SortDirection)
	if err != nil {
		return fmt.Errorf("failed to get list of users: %w", err)
	}

	userCount, err := h.persister.GetUserPersister().Count(userId, email, username)
	if err != nil {
		return fmt.Errorf("failed to get total count of users: %w", err)
	}

	u, _ := url.Parse(fmt.Sprintf("%s://%s%s", c.Scheme(), c.Request().Host, c.Request().RequestURI))

	c.Response().Header().Set("Link", pagination.CreateHeader(u, userCount, request.Page, request.PerPage))
	c.Response().Header().Set("X-Total-Count", strconv.FormatInt(int64(userCount), 10))

	l := make([]admin.User, len(users))
	for i := range users {
		l[i] = admin.FromUserModel(users[i])
	}

	return c.JSON(http.StatusOK, l)
}

func (h *UserHandlerAdmin) Get(c echo.Context) error {
	userId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse userId as uuid").SetInternal(err)
	}

	p := h.persister.GetUserPersister()
	user, err := p.Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	return c.JSON(http.StatusOK, admin.FromUserModel(*user))
}

func (h *UserHandlerAdmin) Create(c echo.Context) error {
	var body admin.CreateUser
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	if len(body.Emails) == 0 && (body.Username == nil || *body.Username == "") {
		return echo.NewHTTPError(http.StatusBadRequest, "at least one of [Emails, Username] must be set")
	}

	// if no userID is provided, create a new one
	if body.ID.IsNil() {
		userId, err := uuid.NewV4()
		if err != nil {
			return fmt.Errorf("failed to create new userId: %w", err)
		}
		body.ID = userId
	}

	// check that only one email is marked as primary
	primaryEmails := 0
	for _, email := range body.Emails {
		if email.IsPrimary {
			primaryEmails++
		}
	}

	if primaryEmails == 0 && len(body.Emails) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "at least one primary email must be provided")
	} else if primaryEmails > 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "only one primary email is allowed")
	}

	err := h.persister.GetConnection().Transaction(func(tx *pop.Connection) error {
		u := models.User{
			ID:        body.ID,
			CreatedAt: body.CreatedAt,
		}

		err := tx.Create(&u)
		if err != nil {
			var pgErr *pgconn.PgError
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" {
					return echo.NewHTTPError(http.StatusConflict, fmt.Errorf("failed to create user with id '%v': %w", u.ID, fmt.Errorf("user already exists")))
				}
			} else if errors.As(err, &mysqlErr) {
				if mysqlErr.Number == 1062 {
					return echo.NewHTTPError(http.StatusConflict, fmt.Errorf("failed to create user with id '%v': %w", u.ID, fmt.Errorf("user already exists")))
				}
			}
			return fmt.Errorf("failed to create user with id '%v': %w", u.ID, err)
		}

		now := time.Now()
		for _, email := range body.Emails {
			emailId, _ := uuid.NewV4()
			mail := models.Email{
				ID:        emailId,
				UserID:    &u.ID,
				Address:   strings.ToLower(email.Address),
				Verified:  email.IsVerified,
				CreatedAt: now,
				UpdatedAt: now,
			}

			err := tx.Create(&mail)
			if err != nil {
				var pgErr *pgconn.PgError
				var mysqlErr *mysql.MySQLError
				if errors.As(err, &pgErr) {
					if pgErr.Code == "23505" {
						return echo.NewHTTPError(http.StatusConflict, fmt.Errorf("failed to create email '%s' for user '%v': %w", mail.Address, u.ID, fmt.Errorf("email already exists")))
					}
				} else if errors.As(err, &mysqlErr) {
					if mysqlErr.Number == 1062 {
						return echo.NewHTTPError(http.StatusConflict, fmt.Errorf("failed to create email '%s' for user '%v': %w", mail.Address, u.ID, fmt.Errorf("email already exists")))
					}
				}
				return fmt.Errorf("failed to create email '%s' for user '%v': %w", mail.Address, u.ID, err)
			}

			if email.IsPrimary {
				primary := models.PrimaryEmail{
					UserID:  u.ID,
					EmailID: mail.ID,
				}
				err := tx.Create(&primary)
				if err != nil {
					return fmt.Errorf("failed to set email '%s' as primary for user '%v': %w", mail.Address, u.ID, err)
				}
			}
		}

		if body.Username != nil {
			username := models.NewUsername(u.ID, *body.Username)
			err = tx.Create(username)
			if err != nil {
				var pgErr *pgconn.PgError
				var mysqlErr *mysql.MySQLError
				if errors.As(err, &pgErr) {
					if pgErr.Code == "23505" {
						return echo.NewHTTPError(http.StatusConflict, fmt.Errorf("failed to create username '%s' for user '%v': %w", username.Username, u.ID, fmt.Errorf("username already exists")))
					}
				} else if errors.As(err, &mysqlErr) {
					if mysqlErr.Number == 1062 {
						return echo.NewHTTPError(http.StatusConflict, fmt.Errorf("failed to create username '%s' for user '%v': %w", username.Username, u.ID, fmt.Errorf("username already exists")))
					}
				}
				return fmt.Errorf("failed to create email '%s' for user '%v': %w", username.Username, u.ID, err)
			}
		}
		return nil
	})

	if httpError, ok := err.(*echo.HTTPError); ok {
		return httpError
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	p := h.persister.GetUserPersister()
	user, err := p.Get(body.ID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	userDto := admin.FromUserModel(*user)

	err = utils.TriggerWebhooks(c, events.UserCreate, userDto)
	if err != nil {
		c.Logger().Warn(err)
	}

	return c.JSON(http.StatusOK, userDto)
}
