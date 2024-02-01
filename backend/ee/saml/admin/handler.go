package admin

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/ee/saml/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"net/http"
)

type SamlAdminHandler interface {
	List(ctx echo.Context) error
	Create(ctx echo.Context) error
	Get(ctx echo.Context) error
	Update(ctx echo.Context) error
	Delete(ctx echo.Context) error
}

type samlAdminHandler struct {
	cfg       *config.Config
	persister persistence.Persister
}

const (
	validateRequestError  = "unable to validate request"
	bindRequestError      = "unable to parse request"
	parseIdError          = "unable to parse provider id: %w"
	providerNotFoundError = "unable to find provider"
)

func NewSamlAdminHandler(cfg *config.Config, persister persistence.Persister) SamlAdminHandler {
	return &samlAdminHandler{
		cfg:       cfg,
		persister: persister,
	}
}

func (s *samlAdminHandler) List(ctx echo.Context) error {
	persister := s.persister.GetSamlIdentityProviderPersister(nil)

	providers, err := persister.List()
	if err != nil {
		ctx.Logger().Error(err)

		return err
	}

	return ctx.JSON(http.StatusOK, providers)
}

func (s *samlAdminHandler) Create(ctx echo.Context) error {
	var createDto dto.SamlCreateProviderRequest
	err := ctx.Bind(&createDto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, bindRequestError).SetInternal(err)
	}

	err = ctx.Validate(&createDto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, validateRequestError).SetInternal(err)
	}

	return s.persister.Transaction(func(tx *pop.Connection) error {
		persister := s.persister.GetSamlIdentityProviderPersister(tx)
		model, err := persister.GetByDomain(createDto.Domain)
		if err != nil {
			ctx.Logger().Error(err)
			return fmt.Errorf("unable to fetch providers from database: %w", err)
		}

		if model != nil {
			return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("a provider with the domain '%s' already exists", createDto.Domain))
		}

		provider, err := createDto.ToModel()
		if err != nil {
			ctx.Logger().Error(err)
			return err
		}

		err = persister.Create(provider, &provider.AttributeMap)
		if err != nil {
			ctx.Logger().Error(err)
			return err
		}

		return ctx.JSON(http.StatusCreated, provider)
	})
}

func (s *samlAdminHandler) Get(ctx echo.Context) error {
	var getDto dto.SamlGetProviderRequest
	err := ctx.Bind(&getDto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, bindRequestError).SetInternal(err)
	}

	err = ctx.Validate(&getDto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, validateRequestError).SetInternal(err)
	}

	providerId, err := uuid.FromString(getDto.ID)
	if err != nil {
		ctx.Logger().Error(err)
		return fmt.Errorf(parseIdError, err)
	}

	persister := s.persister.GetSamlIdentityProviderPersister(nil)

	provider, err := persister.Get(providerId)
	if err != nil {
		ctx.Logger().Error(err)
		return fmt.Errorf("unable to fetch provider from db: %w", err)
	}

	if provider == nil {
		return echo.NewHTTPError(http.StatusNotFound, providerNotFoundError)
	}

	return ctx.JSON(http.StatusOK, provider)
}

func (s *samlAdminHandler) Update(ctx echo.Context) error {
	var updateProviderDto dto.SamlUpdateProviderRequest
	err := ctx.Bind(&updateProviderDto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, bindRequestError).SetInternal(err)
	}

	err = ctx.Validate(&updateProviderDto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, validateRequestError).SetInternal(err)
	}

	return s.persister.Transaction(func(tx *pop.Connection) error {
		persister := s.persister.GetSamlIdentityProviderPersister(nil)
		checkModel, err := persister.GetByDomain(updateProviderDto.Domain)
		if err != nil {
			ctx.Logger().Error(err)
			return fmt.Errorf("unable to fetch providers from database: %w", err)
		}

		providerId, err := uuid.FromString(updateProviderDto.ID)
		if err != nil {
			ctx.Logger().Error(err)
			return fmt.Errorf(parseIdError, err)
		}

		if checkModel != nil && checkModel.ID != providerId {
			return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("a provider with the domain '%s' already exists", updateProviderDto.Domain))
		}

		updateModel, err := persister.Get(providerId)
		if err != nil {
			ctx.Logger().Error(err)
			return err
		}

		if updateModel == nil {
			return echo.NewHTTPError(http.StatusNotFound, providerNotFoundError)
		}

		updateModel = updateProviderDto.UpdateModelFromDto(updateModel)

		err = persister.Update(updateModel)
		if err != nil {
			ctx.Logger().Error(err)
			return err
		}

		return ctx.JSON(http.StatusOK, updateModel)
	})
}

func (s *samlAdminHandler) Delete(ctx echo.Context) error {
	var getDto dto.SamlGetProviderRequest
	err := ctx.Bind(&getDto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, bindRequestError).SetInternal(err)
	}

	err = ctx.Validate(&getDto)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, validateRequestError).SetInternal(err)
	}

	providerId, err := uuid.FromString(getDto.ID)
	if err != nil {
		ctx.Logger().Error(err)
		return fmt.Errorf(parseIdError, err)
	}

	persister := s.persister.GetSamlIdentityProviderPersister(nil)

	provider, err := persister.Get(providerId)
	if err != nil {
		ctx.Logger().Error(err)
		return fmt.Errorf("unable to fetch provider from db: %w", err)
	}

	if provider == nil {
		return echo.NewHTTPError(http.StatusNotFound, providerNotFoundError)
	}

	err = persister.Delete(provider)
	if err != nil {
		ctx.Logger().Error(err)
		return fmt.Errorf("unable to delete provider from db: %w", err)
	}

	return ctx.NoContent(http.StatusNoContent)
}
