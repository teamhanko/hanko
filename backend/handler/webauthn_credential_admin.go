package handler

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/dto/admin"
	"github.com/teamhanko/hanko/backend/persistence"
	"net/http"
)

type WebauthnCredentialAdminHandler interface {
	List(ctx echo.Context) error
	Get(ctx echo.Context) error
	Delete(ctx echo.Context) error
}

type webauthnCredentialAdminHandler struct {
	persister persistence.Persister
}

func NewWebauthnCredentialAdminHandler(persister persistence.Persister) WebauthnCredentialAdminHandler {
	return &webauthnCredentialAdminHandler{
		persister: persister,
	}
}

func (h *webauthnCredentialAdminHandler) List(ctx echo.Context) error {
	listDto, err := loadDto[admin.ListWebauthnCredentialsRequestDto](ctx)
	if err != nil {
		return err
	}

	userID, err := uuid.FromString(listDto.UserID)
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

	credentials, err := h.persister.GetWebauthnCredentialPersister().GetFromUser(userID)
	if err != nil {
		return err
	}

	credentialResponses := make([]dto.WebauthnCredentialResponse, len(credentials))
	for i := range credentials {
		credentialResponses[i] = *dto.FromWebauthnCredentialModel(&credentials[i])
	}

	return ctx.JSON(http.StatusOK, credentialResponses)
}

func (h *webauthnCredentialAdminHandler) Get(ctx echo.Context) error {
	getDto, err := loadDto[admin.GetWebauthnCredentialRequestDto](ctx)
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

	credential, err := h.persister.GetWebauthnCredentialPersister().Get(getDto.WebauthnCredentialID)
	if err != nil {
		return err
	}

	if credential == nil || credential.UserId != userID {
		return echo.NewHTTPError(http.StatusNotFound, "webauthn credential not found")
	}

	return ctx.JSON(http.StatusOK, dto.FromWebauthnCredentialModel(credential))
}

func (h *webauthnCredentialAdminHandler) Delete(ctx echo.Context) error {
	deleteDto, err := loadDto[admin.GetWebauthnCredentialRequestDto](ctx)
	if err != nil {
		return err
	}

	userID, err := uuid.FromString(deleteDto.UserID)
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

	credential, err := h.persister.GetWebauthnCredentialPersister().Get(deleteDto.WebauthnCredentialID)
	if err != nil {
		return err
	}

	if credential == nil || credential.UserId != userID {
		return echo.NewHTTPError(http.StatusNotFound, "webauthn credential not found")
	}

	err = h.persister.GetWebauthnCredentialPersister().Delete(*credential)
	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusNoContent)
}
