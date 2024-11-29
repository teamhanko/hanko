package handler

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/dto/admin"
	"github.com/teamhanko/hanko/backend/persistence"
	"net/http"
)

type OTPAdminHandler interface {
	Get(ctx echo.Context) error
	Delete(ctx echo.Context) error
}

type otpAdminHandler struct {
	persister persistence.Persister
}

func NewOTPAdminHandler(persister persistence.Persister) OTPAdminHandler {
	return &otpAdminHandler{persister: persister}
}

func (h *otpAdminHandler) Get(ctx echo.Context) error {
	getDto, err := loadDto[admin.GetOTPRequestDto](ctx)
	if err != nil {
		return err
	}

	userID, err := uuid.FromString(getDto.UserID)
	if err != nil {
		return fmt.Errorf(parseUserUuidFailureMessage, err)
	}

	userModel, err := h.persister.GetUserPersister().Get(userID)
	if err != nil {
		return err
	}

	if userModel == nil || userModel.OTPSecret == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	return ctx.JSON(http.StatusOK, admin.OTPDto{
		ID:        userModel.OTPSecret.ID,
		CreatedAt: userModel.OTPSecret.CreatedAt,
	})
}

func (h *otpAdminHandler) Delete(ctx echo.Context) error {
	deleteDto, err := loadDto[admin.GetOTPRequestDto](ctx)
	if err != nil {
		return err
	}

	userID, err := uuid.FromString(deleteDto.UserID)
	if err != nil {
		return fmt.Errorf(parseUserUuidFailureMessage, err)
	}

	userModel, err := h.persister.GetUserPersister().Get(userID)
	if err != nil {
		return err
	}

	if userModel == nil || userModel.OTPSecret == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	err = h.persister.GetOTPSecretPersister().Delete(userModel.OTPSecret)
	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusNoContent)
}
