package handler

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto/admin"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"github.com/teamhanko/hanko/backend/webhooks/utils"
	"net/http"
	"strings"
)

type EmailAdminHandler interface {
	List(ctx echo.Context) error
	Create(ctx echo.Context) error

	Get(ctx echo.Context) error
	Delete(ctx echo.Context) error
	SetPrimaryEmail(ctx echo.Context) error
}

type emailAdminHandler struct {
	cfg       *config.Config
	persister persistence.Persister
}

const (
	parseUserUuidFailureMessage   = "failed to parse user_id as uuid: %w"
	fetchUserFromDbFailureMessage = "failed to fetch user from db: %w"
)

func NewEmailAdminHandler(cfg *config.Config, persister persistence.Persister) EmailAdminHandler {
	return &emailAdminHandler{
		cfg:       cfg,
		persister: persister,
	}
}

func loadDto[I admin.EmailRequests](ctx echo.Context) (*I, error) {
	var adminDto I
	err := ctx.Bind(&adminDto)
	if err != nil {
		ctx.Logger().Error(err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = ctx.Validate(adminDto)
	if err != nil {
		ctx.Logger().Error(err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, err)
	}

	return &adminDto, nil
}

func (h *emailAdminHandler) List(ctx echo.Context) error {
	listDto, err := loadDto[admin.ListEmailRequestDto](ctx)
	if err != nil {
		return err
	}

	userId, err := uuid.FromString(listDto.UserId)
	if err != nil {
		return fmt.Errorf(parseUserUuidFailureMessage, err)
	}

	emails, err := h.persister.GetEmailPersister().FindByUserId(userId)
	if err != nil {
		return fmt.Errorf("failed to fetch emails from db: %w", err)
	}

	response := make([]*admin.Email, len(emails))

	for i := range emails {
		response[i] = admin.FromEmailModel(&emails[i])
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h *emailAdminHandler) Create(ctx echo.Context) error {
	createDto, err := loadDto[admin.CreateEmailRequestDto](ctx)
	if err != nil {
		return err
	}

	userId, err := uuid.FromString(createDto.UserId)
	if err != nil {
		return fmt.Errorf(parseUserUuidFailureMessage, err)
	}

	emailCount, err := h.persister.GetEmailPersister().CountByUserId(userId)
	if err != nil {
		return fmt.Errorf("failed to count user emails: %w", err)
	}

	if emailCount >= h.cfg.Email.Limit {
		return echo.NewHTTPError(http.StatusConflict).SetInternal(errors.New("max number of email addresses reached"))
	}

	newEmailAddress := strings.ToLower(createDto.Address)

	email, err := h.persister.GetEmailPersister().FindByAddress(newEmailAddress)
	if err != nil {
		return fmt.Errorf("failed to fetch email from db: %w", err)
	}

	user, err := h.persister.GetUserPersister().Get(userId)
	if err != nil {
		return fmt.Errorf(fetchUserFromDbFailureMessage, err)
	}

	return h.persister.Transaction(func(tx *pop.Connection) error {
		if user == nil {
			return echo.NewHTTPError(http.StatusNotFound).SetInternal(errors.New("user not found"))
		}

		if email != nil {
			// The email address already exists.
			if email.UserID != nil {
				// The email address exists and is assigned to a user already, therefore it can't be created.
				return echo.NewHTTPError(http.StatusBadRequest).SetInternal(errors.New("email address already exists"))
			}

			email.UserID = &user.ID
			err = h.persister.GetEmailPersisterWithConnection(tx).Update(*email)

			if err != nil {
				return fmt.Errorf("failed to update the existing email: %w", err)
			}
		} else {
			email = models.NewEmail(&user.ID, newEmailAddress)
			email.Verified = createDto.IsVerified

			err = h.persister.GetEmailPersisterWithConnection(tx).Create(*email)
			if err != nil {
				return fmt.Errorf("failed to store email to db: %w", err)
			}
		}

		// make email primary if user had no emails prior to email creation
		if len(user.Emails) == 0 {
			primaryEmail := models.NewPrimaryEmail(email.ID, user.ID)
			err = h.persister.GetPrimaryEmailPersisterWithConnection(tx).Create(*primaryEmail)
		}

		utils.NotifyUserChange(ctx, tx, h.persister, events.UserEmailCreate, userId)

		return ctx.JSON(http.StatusCreated, admin.FromEmailModel(email))
	})
}

func (h *emailAdminHandler) Get(ctx echo.Context) error {
	getDto, err := loadDto[admin.GetEmailRequestDto](ctx)
	if err != nil {
		return err
	}

	userId, err := uuid.FromString(getDto.UserId)
	if err != nil {
		return fmt.Errorf(parseUserUuidFailureMessage, err)
	}

	emailId, err := uuid.FromString(getDto.EmailId)
	if err != nil {
		return fmt.Errorf("failed to parse email_id as uuid: %w", err)
	}

	user, err := h.persister.GetUserPersister().Get(userId)
	if err != nil {
		return fmt.Errorf(fetchUserFromDbFailureMessage, err)
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("user with id '%s' was not found", userId))
	}

	fetchedEmail := user.GetEmailById(emailId)
	if fetchedEmail == nil {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Errorf("email with id '%s' was not found", emailId))
	}

	return ctx.JSON(http.StatusOK, admin.FromEmailModel(fetchedEmail))
}

func (h *emailAdminHandler) Delete(ctx echo.Context) error {
	deleteDto, err := loadDto[admin.GetEmailRequestDto](ctx)
	if err != nil {
		return err
	}

	userId, err := uuid.FromString(deleteDto.UserId)
	if err != nil {
		return fmt.Errorf(parseUserUuidFailureMessage, err)
	}

	emailId, err := uuid.FromString(deleteDto.EmailId)
	if err != nil {
		return fmt.Errorf("failed to parse email_id as uuid: %w", err)
	}

	user, err := h.persister.GetUserPersister().Get(userId)
	if err != nil {
		return fmt.Errorf(fetchUserFromDbFailureMessage, err)
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("user with id '%s' was not found", userId))
	}

	emailToBeDeleted := user.GetEmailById(emailId)
	if emailToBeDeleted == nil {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("email with id '%s' was not found", emailId))
	}

	if emailToBeDeleted.IsPrimary() {
		return echo.NewHTTPError(http.StatusConflict).SetInternal(errors.New("primary email can't be deleted"))
	}

	return h.persister.Transaction(func(tx *pop.Connection) error {
		err = h.persister.GetEmailPersisterWithConnection(tx).Delete(*emailToBeDeleted)
		if err != nil {
			return fmt.Errorf("failed to delete email from db: %w", err)
		}

		utils.NotifyUserChange(ctx, tx, h.persister, events.UserEmailDelete, userId)

		return ctx.NoContent(http.StatusNoContent)
	})
}

func (h *emailAdminHandler) SetPrimaryEmail(ctx echo.Context) error {
	primaryDto, err := loadDto[admin.GetEmailRequestDto](ctx)
	if err != nil {
		return err
	}

	userId, err := uuid.FromString(primaryDto.UserId)
	if err != nil {
		return fmt.Errorf(parseUserUuidFailureMessage, err)
	}

	emailId, err := uuid.FromString(primaryDto.EmailId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest).SetInternal(err)
	}

	user, err := h.persister.GetUserPersister().Get(userId)
	if err != nil {
		return fmt.Errorf(fetchUserFromDbFailureMessage, err)
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("user with id '%s' was not found", userId))
	}

	email := user.GetEmailById(emailId)
	if email == nil {
		return echo.NewHTTPError(http.StatusNotFound).SetInternal(errors.New("the email address is not assigned to the current user"))
	}

	if email.IsPrimary() {
		return ctx.NoContent(http.StatusNoContent)
	}

	return h.persister.Transaction(func(tx *pop.Connection) error {
		err := h.makeEmailPrimary(ctx, email, user, tx)
		if err != nil {
			return err
		}

		utils.NotifyUserChange(ctx, tx, h.persister, events.UserEmailPrimary, userId)

		return ctx.NoContent(http.StatusNoContent)
	})
}

func (h *emailAdminHandler) makeEmailPrimary(ctx echo.Context, email *models.Email, user *models.User, tx *pop.Connection) error {
	var primaryEmail *models.PrimaryEmail
	if e := user.Emails.GetPrimary(); e != nil {
		primaryEmail = e.PrimaryEmail
	}

	if primaryEmail == nil {
		primaryEmail = models.NewPrimaryEmail(email.ID, user.ID)
		err := h.persister.GetPrimaryEmailPersisterWithConnection(tx).Create(*primaryEmail)
		if err != nil {
			ctx.Logger().Error(err)
			return fmt.Errorf("failed to store new primary email: %w", err)
		}
	} else {
		primaryEmail.EmailID = email.ID
		err := h.persister.GetPrimaryEmailPersisterWithConnection(tx).Update(*primaryEmail)
		if err != nil {
			ctx.Logger().Error(err)
			return fmt.Errorf("failed to change primary email: %w", err)
		}
	}

	return nil
}
