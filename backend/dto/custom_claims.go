package dto

import (
	"encoding/json"

	zeroLogger "github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/thirdparty"
)

// CustomClaimsFromUserModel returns the flat {claimName: value} JSON view of a user's stored
// custom claims for JWT template processing (see UserJWT.CustomClaims), unwrapping the stored
// {claimName: {value, source}} envelope - source is internal bookkeeping
// (thirdparty.applyCustomClaims's source-scoped clear semantics) and must never reach the JWT.
// Returns nil if the user has no custom claims at all.
func CustomClaimsFromUserModel(customClaims *models.UserCustomClaims) json.RawMessage {
	if customClaims == nil || !customClaims.Claims.Valid || customClaims.Claims.String == "" {
		return nil
	}

	stored := make(map[string]thirdparty.StoredCustomClaim)
	if err := json.Unmarshal([]byte(customClaims.Claims.String), &stored); err != nil {
		// Should never happen - this backend is the only writer of this column and always
		// writes valid JSON. Omit rather than fail token issuance over corrupted data.
		zeroLogger.Warn().Err(err).Str("component", "dto").Str("user_id", customClaims.UserID.String()).
			Msg("could not unmarshal custom claims for JWT, omitting them")
		return nil
	}
	if len(stored) == 0 {
		return nil
	}

	flat := make(map[string]any, len(stored))
	for name, claim := range stored {
		flat[name] = claim.Value
	}

	flatJSON, err := json.Marshal(flat)
	if err != nil {
		zeroLogger.Warn().Err(err).Str("component", "dto").Str("user_id", customClaims.UserID.String()).
			Msg("could not marshal custom claims for JWT, omitting them")
		return nil
	}

	return flatJSON
}
