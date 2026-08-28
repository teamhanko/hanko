package dto

import (
	"encoding/json"
	"strings"

	zeroLogger "github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/thirdparty"
	"github.com/tidwall/gjson"
)

// CustomClaimsJWT represents a user's tenant-defined custom claims for JWT template
// processing. The field is private on purpose, mirroring MetadataJWT, with dedicated methods
// for accessing the data during template processing - but deliberately not an exact mirror
// of MetadataJWT's shape: that type has two accessors (Public/Unsafe) because metadata has a
// public/unsafe split. Custom claims have no such split, so there's a single Get accessor.
type CustomClaimsJWT struct {
	claims json.RawMessage
}

// NewCustomClaimsJWT creates a new CustomClaimsJWT from a flat {claimName: value} JSON raw
// message. Primarily used in tests to construct a CustomClaimsJWT (due to the private field).
func NewCustomClaimsJWT(claims json.RawMessage) *CustomClaimsJWT {
	return &CustomClaimsJWT{claims: claims}
}

// CustomClaimsJWTFromUserModel creates a new CustomClaimsJWT DTO from a UserCustomClaims
// model, unwrapping the stored {claimName: {value, source}} envelope into a flat
// {claimName: value} view - source is internal bookkeeping (thirdparty.applyCustomClaims's
// source-scoped clear semantics) and must never reach the JWT.
func CustomClaimsJWTFromUserModel(customClaims *models.UserCustomClaims) *CustomClaimsJWT {
	if customClaims == nil || len(customClaims.Claims) == 0 {
		return nil
	}

	stored := make(map[string]thirdparty.StoredCustomClaim)
	if err := json.Unmarshal(customClaims.Claims, &stored); err != nil {
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

	return &CustomClaimsJWT{claims: flatJSON}
}

// Get looks up path (a gjson path, joined with "." - see MetadataJWT.Public/Unsafe for the
// same convention) within the flat claims object. With no path, returns the whole object.
func (c *CustomClaimsJWT) Get(path ...string) string {
	if c == nil {
		return ""
	}
	if len(path) < 1 {
		return gjson.GetBytes(c.claims, "@this").String()
	}

	return gjson.GetBytes(c.claims, strings.Join(path, ".")).String()
}

func (c *CustomClaimsJWT) String() string {
	if c == nil {
		return ""
	}
	jsonBytes, _ := json.Marshal(c)
	return string(jsonBytes)
}

func (c *CustomClaimsJWT) MarshalJSON() ([]byte, error) {
	if c.claims == nil {
		return []byte("{}"), nil
	}
	return c.claims, nil
}
