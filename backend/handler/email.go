package handler

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"net/http"
)

type EmailHandler struct {
	persister      persistence.Persister
	cfg            *config.Config
	sessionManager session.Manager
	auditLogger    auditlog.Logger
}

func NewEmailHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, auditLogger auditlog.Logger) (*EmailHandler, error) {
	return &EmailHandler{
		persister:      persister,
		cfg:            cfg,
		sessionManager: sessionManager,
		auditLogger:    auditLogger,
	}, nil
}

func (h *EmailHandler) List(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}

	userId, err := uuid.FromString(sessionToken.Subject())
	if err != nil {
		return fmt.Errorf("failed to parse subject as uuid: %w", err)
	}

	emails, err := h.persister.GetEmailPersister().FindByUserId(userId)
	if err != nil {
		return fmt.Errorf("failed to fetch emails from db: %w", err)
	}

	emailListResponse := make([]*dto.EmailResponse, len(emails))

	for i := range emails {
		emailListResponse[i] = dto.FromEmailModel(&emails[i])
	}

	return c.JSON(http.StatusOK, emailListResponse)
}

func (h *EmailHandler) Create(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}

	userId, err := uuid.FromString(sessionToken.Subject())
	if err != nil {
		return fmt.Errorf("failed to parse subject as uuid: %w", err)
	}

	var body dto.EmailCreateRequest

	err = (&echo.DefaultBinder{}).BindBody(c, &body)
	if err != nil {
		return dto.ToHttpError(err)
	}

	existingEmail, err := h.persister.GetEmailPersister().FindByAddress(body.Address)
	if err != nil {
		return fmt.Errorf("failed to fetch email from db: %w", err)
	}

	if existingEmail != nil {
		return dto.NewHTTPError(http.StatusConflict).SetInternal(errors.New("email address already exists"))
	}

	email := models.NewEmail(userId, body.Address)

	err = h.persister.GetEmailPersister().Create(*email)
	if err != nil {
		return errors.New("failed to store email to db")
	}

	return c.JSON(http.StatusCreated, nil)
}

func (h *EmailHandler) Update(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}

	userId, err := uuid.FromString(sessionToken.Subject())
	if err != nil {
		return fmt.Errorf("failed to parse subject as uuid: %w", err)
	}

	id := c.Param("id")
	emailId, err := uuid.FromString(id)
	if err != nil {
		return dto.NewHTTPError(http.StatusBadRequest).SetInternal(err)
	}

	var body dto.EmailUpdateRequest

	err = (&echo.DefaultBinder{}).BindBody(c, &body)
	if err != nil {
		return dto.ToHttpError(err)
	}

	user, err := h.persister.GetUserPersister().Get(userId)
	if err != nil {
		return fmt.Errorf("failed to fetch user from db: %w", err)
	}

	email := user.GetEmailById(emailId)
	if email == nil {
		return dto.NewHTTPError(http.StatusNotFound).SetInternal(errors.New("the user does not have an email with the specified emailId"))
	}

	if body.IsPrimary != nil && *body.IsPrimary != email.IsPrimary() {
		primaryEmail := user.GetPrimaryEmail()

		if primaryEmail == nil {
			return errors.New("user has no primary email")
		}

		if h.cfg.Flow.RequireEmailVerification && !primaryEmail.Verified {
			return dto.NewHTTPError(http.StatusConflict).SetInternal(errors.New("email address must be verified to be set as primary email"))
		}

		// Mark email address with specified emailId as primary email address
		primaryEmail.PrimaryEmail.EmailID = emailId

		err = h.persister.GetPrimaryEmailPersister().Update(*primaryEmail.PrimaryEmail)
		if err != nil {
			return fmt.Errorf("failed to store updated primary email to db: %w", err)
		}
	}

	return c.JSON(http.StatusNoContent, nil)
}

func (h *EmailHandler) Delete(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}

	userId, err := uuid.FromString(sessionToken.Subject())
	if err != nil {
		return fmt.Errorf("failed to parse subject as uuid: %w", err)
	}

	emailId, err := uuid.FromString(c.Param("id"))

	user, err := h.persister.GetUserPersister().Get(userId)
	if err != nil {
		return fmt.Errorf("failed to fetch user from db: %w", err)
	}

	emailToBeDeleted := user.GetEmailById(emailId)
	if emailToBeDeleted == nil {
		return errors.New("email with given emailId not available")
	}

	if emailToBeDeleted.IsPrimary() {
		return dto.NewHTTPError(http.StatusConflict).SetInternal(errors.New("primary email can't be deleted"))
	}

	err = h.persister.GetEmailPersister().Delete(*emailToBeDeleted)
	if err != nil {
		return fmt.Errorf("failed to delete email from db: %w", err)
	}

	return c.JSON(http.StatusNoContent, nil)
}
