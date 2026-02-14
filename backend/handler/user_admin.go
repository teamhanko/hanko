package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgconn"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/dto/admin"
	"github.com/teamhanko/hanko/backend/v2/pagination"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/utils"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
	webhookUtils "github.com/teamhanko/hanko/backend/v2/webhooks/utils"
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

	err = h.persister.Transaction(func(tx *pop.Connection) error {
		p := h.persister.GetUserPersisterWithConnection(tx)
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

		err = webhookUtils.TriggerWebhooks(c, tx, events.UserDelete, admin.FromUserModel(*user))
		if err != nil {
			c.Logger().Warn(err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

type UserListRequest struct {
	PerPage       int    `query:"per_page"`
	Page          int    `query:"page"`
	Email         string `query:"email"`
	UserID        string `query:"user_id"`
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

	var userIDs []uuid.UUID
	if request.UserID != "" {
		for _, userIDString := range strings.Split(request.UserID, ",") {
			userID, err := uuid.FromString(userIDString)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "failed to parse user_id as uuid").SetInternal(err)
			}
			userIDs = append(userIDs, userID)
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

	users, err := h.persister.GetUserPersister().List(request.Page, request.PerPage, userIDs, email, username, request.SortDirection)
	if err != nil {
		return fmt.Errorf("failed to get list of users: %w", err)
	}

	userCount, err := h.persister.GetUserPersister().Count(userIDs, email, username)
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

	err = webhookUtils.TriggerWebhooks(c, h.persister.GetConnection(), events.UserCreate, userDto)
	if err != nil {
		c.Logger().Warn(err)
	}

	return c.JSON(http.StatusOK, userDto)
}

// OptionalString represents a PATCH-able string field with 3 states:
// - not present in JSON => Present=false (no change)
// - present with string => Present=true, Value!=nil (set)
// - present with null   => Present=true, Value==nil (clear)
type OptionalString struct {
	Present bool
	Value   *string
}

func (o *OptionalString) UnmarshalJSON(b []byte) error {
	o.Present = true

	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		o.Value = nil
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("expected string or null: %w", err)
	}

	o.Value = &s
	return nil
}

type PatchUserAdminRequest struct {
	Username   OptionalString `json:"username"`
	Name       OptionalString `json:"name"`
	GivenName  OptionalString `json:"given_name"`
	FamilyName OptionalString `json:"family_name"`
	Picture    OptionalString `json:"picture"`
}

func (h *UserHandlerAdmin) Patch(c echo.Context) error {
	userId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse userId as uuid").SetInternal(err)
	}

	var body PatchUserAdminRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	// Empty/whitespace-only strings are invalid (`null` is used to clear).
	normalizeOptionalString := func(field string, v OptionalString, lower bool) (OptionalString, error) {
		if !v.Present || v.Value == nil {
			return v, nil
		}

		s := strings.TrimSpace(*v.Value)
		if s == "" {
			return v, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%s must be a non-empty string or null", field))
		}
		if lower {
			s = strings.ToLower(s)
		}
		v.Value = &s
		return v, nil
	}

	body.Username, err = normalizeOptionalString("username", body.Username, true)
	if err != nil {
		return err
	}
	body.Name, err = normalizeOptionalString("name", body.Name, false)
	if err != nil {
		return err
	}
	body.GivenName, err = normalizeOptionalString("given_name", body.GivenName, false)
	if err != nil {
		return err
	}
	body.FamilyName, err = normalizeOptionalString("family_name", body.FamilyName, false)
	if err != nil {
		return err
	}
	body.Picture, err = normalizeOptionalString("picture", body.Picture, false)
	if err != nil {
		return err
	}

	if body.Picture.Present && body.Picture.Value != nil {
		if reason := utils.ValidatePictureURL(*body.Picture.Value); reason != "" {
			return echo.NewHTTPError(http.StatusBadRequest, "picture must be a valid http(s) URL or null")
		}
	}

	err = h.persister.Transaction(func(tx *pop.Connection) error {
		userPersister := h.persister.GetUserPersisterWithConnection(tx)
		usernamePersister := h.persister.GetUsernamePersisterWithConnection(tx)

		user, err := userPersister.Get(userId)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		if user == nil {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}

		changed := false

		applyNullsString := func(dst *nulls.String, in OptionalString) {
			if !in.Present {
				return
			}
			if in.Value == nil {
				*dst = nulls.String{} // NULL
			} else {
				*dst = nulls.NewString(*in.Value)
			}
			changed = true
		}

		applyNullsString(&user.Name, body.Name)
		applyNullsString(&user.GivenName, body.GivenName)
		applyNullsString(&user.FamilyName, body.FamilyName)
		applyNullsString(&user.Picture, body.Picture)

		// Username is currently still stored in its own table/record.
		if body.Username.Present {
			if body.Username.Value == nil {
				if user.Username != nil {
					if err := usernamePersister.Delete(user.Username); err != nil {
						return fmt.Errorf("failed to delete username: %w", err)
					}
					user.DeleteUsername()
					changed = true
				}
			} else {
				newUsername := *body.Username.Value

				validNewUsername := regexp.MustCompile(`^\w+$`).MatchString(newUsername)

				if !validNewUsername {
					return echo.NewHTTPError(http.StatusBadRequest, "username is invalid")
				}

				dup, err := usernamePersister.GetByName(newUsername)
				if err != nil {
					return fmt.Errorf("failed to check duplicate username: %w", err)
				}
				if dup != nil && dup.UserId != user.ID {
					return echo.NewHTTPError(http.StatusConflict, "username already exists")
				}

				if user.Username == nil {
					usernameModel := models.NewUsername(user.ID, newUsername)
					if err := usernamePersister.Create(*usernameModel); err != nil {
						return fmt.Errorf("failed to create username: %w", err)
					}
					user.SetUsername(usernameModel)
					changed = true
				} else if user.Username.Username != newUsername {
					user.Username.Username = newUsername
					user.Username.UpdatedAt = time.Now()
					if err := usernamePersister.Update(user.Username); err != nil {
						return fmt.Errorf("failed to update username: %w", err)
					}
					changed = true
				}
			}
		}

		if !changed {
			return nil
		}

		user.UpdatedAt = time.Now()
		if err := userPersister.Update(*user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	user, err := h.persister.GetUserPersister().Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	return c.JSON(http.StatusOK, admin.FromUserModel(*user))
}
