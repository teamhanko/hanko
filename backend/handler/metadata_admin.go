package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/dto/admin"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/tidwall/gjson"
)

type MetadataAdminHandler struct {
	persister persistence.Persister
}

func NewMetadataAdminHandler(persister persistence.Persister) *MetadataAdminHandler {
	return &MetadataAdminHandler{
		persister: persister,
	}
}

func (h *MetadataAdminHandler) GetMetadata(c echo.Context) error {
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

	metadataModel, err := h.persister.GetUserMetadataPersister().Get(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not fetch metadata").SetInternal(err)
	}

	response := admin.NewMetadata(metadataModel)

	if response == nil {
		return c.NoContent(http.StatusNoContent)
	}
	return c.JSON(http.StatusOK, response)
}

func (h *MetadataAdminHandler) PatchMetadata(c echo.Context) error {
	userID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	patchMetadataRequest, err := loadDto[admin.PatchMetadataRequest](c)
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

	currentMetadataModel, err := h.persister.GetUserMetadataPersister().Get(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not fetch metadata").SetInternal(err)
	}

	_, err = h.applyMetadataPatch(currentMetadataModel, patchMetadataRequest)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not patch metadata").SetInternal(err)
	}

	err = h.persister.GetUserMetadataPersister().Update(currentMetadataModel)
	if err != nil {
		if persistence.IsMetadataLimitExceededError(err) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error()).
				SetInternal(errors.Unwrap(err))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "could not save metadata").
			SetInternal(err)
	}

	response := admin.NewMetadata(currentMetadataModel)

	if response == nil {
		return c.NoContent(http.StatusNoContent)
	}
	return c.JSON(http.StatusOK, response)
}

func (h *MetadataAdminHandler) applyMetadataPatch(currentMetadataModel *models.UserMetadata, patchMetadataRequest *admin.PatchMetadataRequest) ([]byte, error) {
	if patchMetadataRequest.Metadata.Raw == "null" {
		currentMetadataModel.Public = nulls.String{}
		currentMetadataModel.Private = nulls.String{}
		currentMetadataModel.Unsafe = nulls.String{}
		return []byte("{}"), nil
	}

	currentMetadataJSON, err := h.buildCurrentMetadataJSON(currentMetadataModel)
	if err != nil {
		return nil, err
	}

	patchedMetadataBytes, err := jsonpatch.MergePatch(
		[]byte(currentMetadataJSON),
		[]byte(patchMetadataRequest.Metadata.String()),
	)
	if err != nil {
		return nil, fmt.Errorf("could not apply merge patch: %w", err)
	}

	if !json.Valid(patchedMetadataBytes) {
		return nil, fmt.Errorf("invalid metadata JSON after applying merge patch")
	}

	if err = h.updateMetadataModel(currentMetadataModel, patchedMetadataBytes); err != nil {
		return nil, fmt.Errorf("could not update user metadata model: %w", err)
	}

	return patchedMetadataBytes, nil
}

func (h *MetadataAdminHandler) buildCurrentMetadataJSON(metadata *models.UserMetadata) (string, error) {
	result := make(map[string]json.RawMessage)

	if metadata.Public.Valid && metadata.Public.String != "{}" {
		result["public_metadata"] = json.RawMessage(metadata.Public.String)
	}

	if metadata.Private.Valid && metadata.Private.String != "{}" {
		result["private_metadata"] = json.RawMessage(metadata.Private.String)
	}

	if metadata.Unsafe.Valid && metadata.Unsafe.String != "{}" {
		result["unsafe_metadata"] = json.RawMessage(metadata.Unsafe.String)
	}

	if len(result) == 0 {
		return "{}", nil
	}

	bytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("could not build JSON for current metadata: %w", err)
	}

	return string(bytes), nil
}

func (h *MetadataAdminHandler) updateMetadataModel(metadata *models.UserMetadata, patchedBytes []byte) error {

	metadataTypes := map[string]*nulls.String{
		"public_metadata":  &metadata.Public,
		"private_metadata": &metadata.Private,
		"unsafe_metadata":  &metadata.Unsafe,
	}

	for key, target := range metadataTypes {
		if gjson.GetBytes(patchedBytes, key).Exists() {
			value := gjson.GetBytes(patchedBytes, key).String()
			if !json.Valid([]byte(value)) {
				return fmt.Errorf("invalid JSON for %s", key)
			}
			*target = nulls.String{
				Valid:  true,
				String: value,
			}
		} else {
			*target = nulls.String{}
		}
	}

	return nil
}
