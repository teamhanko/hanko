package handler

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/dto/admin"
	"github.com/teamhanko/hanko/backend/v2/flow_api/services"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"net/http"
)

type PasswordAdminHandler interface {
	Get(ctx echo.Context) error
	Create(ctx echo.Context) error
	Update(ctx echo.Context) error
	Delete(ctx echo.Context) error
}

type passwordAdminHandler struct {
	persister       persistence.Persister
	passwordService services.Password
}

func NewPasswordAdminHandler(persister persistence.Persister) PasswordAdminHandler {
	return &passwordAdminHandler{
		persister:       persister,
		passwordService: services.NewPasswordService(persister),
	}
}

func (h *passwordAdminHandler) Get(ctx echo.Context) error {
	getDto, err := loadDto[admin.GetPasswordCredentialRequestDto](ctx)
	if err != nil {
		return err
	}

	userID, err := uuid.FromString(getDto.UserID)
	if err != nil {
		return fmt.Errorf(parseUserUuidFailureMessage, err)
	}

	user, err := h.persister.GetUserPersister().Get(userID)
	if err != nil {
		return err
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	credential, err := h.persister.GetPasswordCredentialPersister().GetByUserID(userID)
	if err != nil {
		return err
	}

	if credential == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	dto := admin.PasswordCredential{
		ID:        credential.ID,
		CreatedAt: credential.CreatedAt,
		UpdatedAt: credential.UpdatedAt,
	}

	return ctx.JSON(http.StatusOK, dto)
}

func (h *passwordAdminHandler) Create(ctx echo.Context) error {
	createDto, err := loadDto[admin.CreateOrUpdatePasswordCredentialRequestDto](ctx)
	if err != nil {
		return err
	}

	userID, err := uuid.FromString(createDto.UserID)
	if err != nil {
		return fmt.Errorf(parseUserUuidFailureMessage, err)
	}

	user, err := h.persister.GetUserPersister().Get(userID)
	if err != nil {
		return err
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	existingCredential, err := h.persister.GetPasswordCredentialPersister().GetByUserID(userID)
	if err != nil {
		return err
	}

	if existingCredential != nil {
		return echo.NewHTTPError(http.StatusConflict)
	}

	err = h.passwordService.CreatePassword(h.persister.GetConnection(), userID, createDto.Password)
	if err != nil {
		return err
	}

	credential, err := h.persister.GetPasswordCredentialPersister().GetByUserID(userID)
	if err != nil {
		return err
	}

	dto := admin.PasswordCredential{
		ID:        credential.ID,
		CreatedAt: credential.CreatedAt,
		UpdatedAt: credential.UpdatedAt,
	}

	return ctx.JSON(http.StatusOK, dto)
}

func (h *passwordAdminHandler) Update(ctx echo.Context) error {
	updateDto, err := loadDto[admin.CreateOrUpdatePasswordCredentialRequestDto](ctx)
	if err != nil {
		return err
	}

	userID, err := uuid.FromString(updateDto.UserID)
	if err != nil {
		return fmt.Errorf(parseUserUuidFailureMessage, err)
	}

	user, err := h.persister.GetUserPersister().Get(userID)
	if err != nil {
		return err
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	credential, err := h.persister.GetPasswordCredentialPersister().GetByUserID(userID)
	if err != nil {
		return err
	}

	if credential == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	err = h.passwordService.UpdatePassword(h.persister.GetConnection(), credential, updateDto.Password)
	if err != nil {
		return err
	}

	credential, err = h.persister.GetPasswordCredentialPersister().GetByUserID(userID)
	if err != nil {
		return err
	}

	dto := admin.PasswordCredential{
		ID:        credential.ID,
		CreatedAt: credential.CreatedAt,
		UpdatedAt: credential.UpdatedAt,
	}

	return ctx.JSON(http.StatusOK, dto)
}

func (h *passwordAdminHandler) Delete(ctx echo.Context) error {
	getDto, err := loadDto[admin.GetPasswordCredentialRequestDto](ctx)
	if err != nil {
		return err
	}

	userID, err := uuid.FromString(getDto.UserID)
	if err != nil {
		return fmt.Errorf(parseUserUuidFailureMessage, err)
	}

	user, err := h.persister.GetUserPersister().Get(userID)
	if err != nil {
		return err
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	credential, err := h.persister.GetPasswordCredentialPersister().GetByUserID(userID)
	if err != nil {
		return err
	}

	if credential == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	err = h.persister.GetPasswordCredentialPersister().Delete(*credential)
	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusNoContent)
}
