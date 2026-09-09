package admin

import (
	"encoding/json"
	"fmt"

	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/thirdparty"
)

// CustomClaims is the Admin API's flat view of a user's custom claims - only the value of
// each entry, never the internal source tag (thirdparty.StoredCustomClaim.Source) that
// governs which connection may clear it.
type CustomClaims map[string]any

// NewCustomClaims unwraps the stored {claimName: {value, source}} envelope into a flat
// {claimName: value} view. Returns nil (leading to 204 No Content, mirroring metadata) if
// the user has no custom claims at all.
func NewCustomClaims(model *models.UserCustomClaims) (CustomClaims, error) {
	if !model.Claims.Valid || model.Claims.String == "" {
		return nil, nil
	}

	stored := make(map[string]thirdparty.StoredCustomClaim)
	if err := json.Unmarshal([]byte(model.Claims.String), &stored); err != nil {
		return nil, fmt.Errorf("could not unmarshal custom claims: %w", err)
	}

	if len(stored) == 0 {
		return nil, nil
	}

	result := make(CustomClaims, len(stored))
	for name, claim := range stored {
		result[name] = claim.Value
	}

	return result, nil
}
