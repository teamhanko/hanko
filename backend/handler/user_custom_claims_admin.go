package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/context"
	"github.com/teamhanko/hanko/backend/v3/dto/admin"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/thirdparty"
	"github.com/tidwall/gjson"
)

type UserCustomClaimsAdminHandler struct {
	persister persistence.Persister
}

func NewUserCustomClaimsAdminHandler(persister persistence.Persister) *UserCustomClaimsAdminHandler {
	return &UserCustomClaimsAdminHandler{persister: persister}
}

func (h *UserCustomClaimsAdminHandler) GetCustomClaims(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	userID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	userExists, err := h.persister.GetConnection().Where("id = ?", userID).Exists(&models.User{ID: userID})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not fetch user").SetInternal(err)
	}
	if !userExists {
		return echo.NewHTTPError(http.StatusNotFound, "user not found").SetInternal(err)
	}

	customClaimsModel, err := h.persister.GetUserCustomClaimsPersister().Get(userID, tenant.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not fetch custom claims").SetInternal(err)
	}

	response, err := admin.NewCustomClaims(customClaimsModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not build custom claims response").SetInternal(err)
	}

	if response == nil {
		return c.NoContent(http.StatusNoContent)
	}
	return c.JSON(http.StatusOK, response)
}

func (h *UserCustomClaimsAdminHandler) PatchCustomClaims(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	userID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	patchRequest, err := loadDto[admin.PatchCustomClaimsRequest](c)
	if err != nil {
		return err
	}

	userExists, err := h.persister.GetConnection().Where("id = ?", userID).Exists(&models.User{ID: userID})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not fetch user").SetInternal(err)
	}
	if !userExists {
		return echo.NewHTTPError(http.StatusNotFound, "user not found").SetInternal(err)
	}

	customClaimsModel, err := h.persister.GetUserCustomClaimsPersister().Get(userID, tenant.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not fetch custom claims").SetInternal(err)
	}

	if err := applyCustomClaimsPatch(tenant.Config.CustomClaims.Definitions, customClaimsModel, patchRequest); err != nil {
		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			return httpErr
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "could not patch custom claims").SetInternal(err)
	}

	if err := h.persister.GetUserCustomClaimsPersister().Update(customClaimsModel); err != nil {
		if persistence.IsCustomClaimsLimitExceededError(err) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(errors.Unwrap(err))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "could not save custom claims").SetInternal(err)
	}

	response, err := admin.NewCustomClaims(customClaimsModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not build custom claims response").SetInternal(err)
	}

	if response == nil {
		return c.NoContent(http.StatusNoContent)
	}
	return c.JSON(http.StatusOK, response)
}

// applyCustomClaimsPatch applies an RFC 7386 JSON merge patch to model.Claims. A null body
// clears every stored claim, mirroring PatchMetadata's null-wipes-everything precedent. A
// null value for a specific key clears only that key. Every non-null key must already be
// declared in defs and match its declared type exactly (no coercion, unlike
// thirdparty.ResolveCustomClaims: an admin authoring JSON directly is expected to send the
// correct native JSON type, not a loosely-typed IdP-style string) - a violation returns an
// *echo.HTTPError(400) directly, which the caller unwraps and returns as-is.
func applyCustomClaimsPatch(defs config.CustomClaimDefinitions, model *models.UserCustomClaims, patch *admin.PatchCustomClaimsRequest) error {
	if patch.Claims.Raw == "null" {
		model.Claims = json.RawMessage("{}")
		return nil
	}

	stored := make(map[string]thirdparty.StoredCustomClaim)
	if len(model.Claims) > 0 {
		if err := json.Unmarshal(model.Claims, &stored); err != nil {
			return fmt.Errorf("could not unmarshal existing custom claims: %w", err)
		}
	}

	var patchErr error
	patch.Claims.ForEach(func(key, value gjson.Result) bool {
		claimName := key.String()

		definition, declared := defs[claimName]
		if !declared {
			patchErr = echo.NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("custom claim %q is not declared in custom_claims.definitions", claimName))
			return false
		}

		if value.Type == gjson.Null {
			delete(stored, claimName)
			return true
		}

		coerced, err := matchCustomClaimType(definition, value)
		if err != nil {
			patchErr = echo.NewHTTPError(http.StatusBadRequest, err.Error())
			return false
		}

		stored[claimName] = thirdparty.StoredCustomClaim{Value: coerced, Source: "admin"}
		return true
	})
	if patchErr != nil {
		return patchErr
	}

	claimsJSON, err := json.Marshal(stored)
	if err != nil {
		return fmt.Errorf("could not marshal custom claims: %w", err)
	}
	model.Claims = claimsJSON

	return nil
}

// matchCustomClaimType requires value's JSON type to exactly match the declared type -
// deliberately strict, unlike thirdparty.ResolveCustomClaims's lenient IdP-attribute
// coercion (see applyCustomClaimsPatch's doc comment for why).
func matchCustomClaimType(definition config.CustomClaimDefinition, value gjson.Result) (any, error) {
	switch definition.Type {
	case config.CustomClaimTypeString:
		if value.Type != gjson.String {
			return nil, fmt.Errorf("custom claim %q must be a string", definition.Name)
		}
		return value.String(), nil
	case config.CustomClaimTypeNumber:
		if value.Type != gjson.Number {
			return nil, fmt.Errorf("custom claim %q must be a number", definition.Name)
		}
		return value.Num, nil
	case config.CustomClaimTypeBoolean:
		if value.Type != gjson.True && value.Type != gjson.False {
			return nil, fmt.Errorf("custom claim %q must be a boolean", definition.Name)
		}
		return value.Bool(), nil
	case config.CustomClaimTypeStringList:
		if !value.IsArray() {
			return nil, fmt.Errorf("custom claim %q must be an array of strings", definition.Name)
		}
		list := make([]string, 0)
		for _, element := range value.Array() {
			if element.Type != gjson.String {
				return nil, fmt.Errorf("custom claim %q must be an array of strings", definition.Name)
			}
			list = append(list, element.String())
		}
		return list, nil
	default:
		return nil, fmt.Errorf("custom claim %q has an unknown declared type %q", definition.Name, definition.Type)
	}
}
