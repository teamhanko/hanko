package thirdparty

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	zeroLogger "github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

// ResolveCustomClaims coerces mapped provider attributes into the tenant's declared claim
// types. managed is every claim name this connection is responsible for - every key in
// src.Mapping that's also currently declared in defs - whether or not a value was actually
// resolved this time. Callers (applyCustomClaims) use managed to decide which claims this
// connection may clear on absence.
//
// Never fails: an undeclared claim name, a missing attribute, or a value that doesn't coerce
// to the declared type are all logged and skipped, not returned as an error - a malformed or
// unmapped attribute must never fail a login.
func ResolveCustomClaims(defs config.CustomClaimDefinitions, src *CustomClaimSource) (values map[string]any, managed []string) {
	if src == nil {
		return nil, nil
	}

	values = make(map[string]any)

	for claimName, attributeName := range src.Mapping {
		definition, declared := defs[claimName]
		if !declared {
			zeroLogger.Warn().
				Str("component", "thirdparty").
				Str("operation", "resolve_custom_claims").
				Str("claim", claimName).
				Msg("skipping custom claim mapping: not declared in custom_claims.definitions")
			continue
		}
		managed = append(managed, claimName)

		raw, present := src.Attributes[attributeName]
		if !present || raw == nil {
			continue
		}

		value, ok := coerceCustomClaim(definition.Type, raw)
		if !ok {
			zeroLogger.Warn().
				Str("component", "thirdparty").
				Str("operation", "resolve_custom_claims").
				Str("claim", claimName).
				Str("type", definition.Type).
				Msg("skipping custom claim: value could not be coerced to the declared type")
			continue
		}
		values[claimName] = value
	}

	return values, managed
}

// coerceCustomClaim reduces an already-flattened scalar/list value into the shape the
// declared type expects. Returns ok=false on any parse failure or unsupported input shape.
func coerceCustomClaim(claimType string, raw any) (value any, ok bool) {
	flattened, ok := flattenCustomClaimValue(raw)
	if !ok || len(flattened) == 0 {
		return nil, false
	}

	switch claimType {
	case config.CustomClaimTypeString:
		return flattened[0], true
	case config.CustomClaimTypeNumber:
		n, err := strconv.ParseFloat(flattened[0], 64)
		if err != nil {
			return nil, false
		}
		return n, true
	case config.CustomClaimTypeBoolean:
		b, err := strconv.ParseBool(flattened[0])
		if err != nil {
			return nil, false
		}
		return b, true
	case config.CustomClaimTypeStringList:
		return flattened, true
	default:
		return nil, false
	}
}

// flattenCustomClaimValue normalizes a raw provider attribute value - a SAML attribute is
// always string or []string (already flat); an OIDC claim is arbitrary decoded JSON, so it
// may also be bool, float64, or []interface{} - into a flat []string, one entry per scalar
// value. A JSON object, or a list containing one, is deliberately rejected (ok=false) rather
// than stringified: see CustomClaimDefinition.Type's doc comment for why custom claims are
// scalar/list-of-scalar only.
func flattenCustomClaimValue(raw any) (values []string, ok bool) {
	switch v := raw.(type) {
	case nil:
		return nil, true
	case string:
		return []string{v}, true
	case []string:
		return v, true
	case bool:
		return []string{strconv.FormatBool(v)}, true
	case float64:
		return []string{strconv.FormatFloat(v, 'g', -1, 64)}, true
	case []interface{}:
		flattened := make([]string, 0, len(v))
		for _, element := range v {
			elementValues, elementOk := flattenCustomClaimValue(element)
			if !elementOk || len(elementValues) != 1 {
				// a nested list or an object element - reject the whole value.
				return nil, false
			}
			flattened = append(flattened, elementValues[0])
		}
		return flattened, true
	default:
		// includes map[string]interface{} (a JSON object) and any other unrecognized shape.
		return nil, false
	}
}

// StoredCustomClaim is the envelope each entry in user_custom_claims.claims is stored as,
// keyed by claim name. Source is the writing connection's identifier ("saml:<provider_id>",
// "third_party:<provider_id>" - covering both OIDC and plain OAuth2 custom providers, since
// CustomThirdPartyProvider isn't necessarily OIDC-conformant). Tracked so applyCustomClaims
// only ever lets the connection that set a claim clear it again - see applyCustomClaims's doc
// comment for why that matters.
type StoredCustomClaim struct {
	Value  any    `json:"value"`
	Source string `json:"source"`
}

// customClaimConnectionSource builds the source identifier applyCustomClaims tags every
// value it writes with.
//
// This relies on providerID being a stable, permanent identifier for the connection. For
// SAML that's the IdP's own issuer URL (already the key SAML identity linking/dedup relies
// on elsewhere - not something introduced here). For a custom third-party provider, it's
// derived from the admin-chosen key under third_party.custom_providers (see that field's
// doc comment on config.ThirdParty) - if an admin renames that key, claims previously set by
// that connection become orphaned: still stored, but no longer recognized as "owned" by the
// connection under its new identity, so it can no longer clear them itself.
func customClaimConnectionSource(providerID string, isSaml bool) string {
	if isSaml {
		return "saml:" + providerID
	}
	return "third_party:" + providerID
}

// customClaimValueEqual compares a freshly resolved claim value against an already-stored one.
// Both are compared via their JSON representation rather than Go's == - resolvedValues holds
// []string for a string_list claim (coerceCustomClaim's output), while a stored value of the
// same list, once round-tripped through json.Unmarshal into an `any`, comes back as
// []interface{}; those have different dynamic types, so == would either report a spurious
// "changed" (different types with same content) or panic (two values of the same uncomparable
// slice type) - comparing marshaled JSON sidesteps both.
func customClaimValueEqual(a, b any) bool {
	aJSON, aErr := json.Marshal(a)
	bJSON, bErr := json.Marshal(b)
	if aErr != nil || bErr != nil {
		return false
	}
	return string(aJSON) == string(bJSON)
}

// mergeCustomClaims applies resolvedValues onto stored in place, per applyCustomClaims's
// refresh semantics. Split out from applyCustomClaims so this decision logic is testable
// without a DB.
//
// managed (from ResolveCustomClaims) is every claim name this connection maps, whether or not
// it resolved a value this time - it's what tells apart "mapped but the IdP sent nothing this
// login" (may need clearing) from "this connection doesn't map this claim at all" (never
// touched, no matter what's stored). A managed claim missing from resolvedValues is a clear
// candidate; anything not in managed is invisible to this call entirely.
//
// wrote reports whether stored was mutated at all, including a source-only handoff (another
// connection re-asserting a value this claim already had) - that must still persist so
// ownership stays correct for future clears. valueChanged reports whether a claim's value
// itself appeared, disappeared, or changed - false for a source-only handoff.
func mergeCustomClaims(stored map[string]StoredCustomClaim, resolvedValues map[string]any, managed []string, source string) (wrote bool, valueChanged bool) {
	for _, claimName := range managed {
		if value, ok := resolvedValues[claimName]; ok {
			existing, exists := stored[claimName]
			sameValue := exists && customClaimValueEqual(existing.Value, value)
			if exists && sameValue && existing.Source == source {
				continue
			}
			stored[claimName] = StoredCustomClaim{Value: value, Source: source}
			wrote = true
			if !sameValue {
				valueChanged = true
			}
			continue
		}

		if existing, exists := stored[claimName]; exists && existing.Source == source {
			delete(stored, claimName)
			wrote = true
			valueChanged = true
		}
	}
	return wrote, valueChanged
}

// applyCustomClaims resolves userData.CustomClaimSource and merges it into the user's
// persisted custom claims. Last-one-wins, any source; a claim found no value for is cleared
// only if this same connection was its stored source.
//
// record is non-nil whenever anything was written - including a source-only handoff (another
// connection re-asserting a value this claim already had), which still needs persisting so
// ownership stays correct for future clears. valueChanged is true only when a claim's value
// itself appeared, disappeared, or changed - gate a user.update webhook on this, not on
// record != nil, so re-asserting an unchanged value never fires one.
func applyCustomClaims(tx *pop.Connection, p persistence.Persister, cfg *config.TenantConfig, userData *UserData, source string, userID uuid.UUID, tenantID uuid.UUID) (record *models.UserCustomClaims, valueChanged bool, err error) {
	if userData.CustomClaimSource == nil {
		return nil, false, nil
	}

	resolvedValues, managed := ResolveCustomClaims(cfg.CustomClaims.Definitions, userData.CustomClaimSource)
	if len(managed) == 0 {
		return nil, false, nil
	}

	persister := p.GetUserCustomClaimsPersisterWithConnection(tx)
	record, err = persister.Get(userID, tenantID)
	if err != nil {
		return nil, false, fmt.Errorf("could not get user custom claims: %w", err)
	}

	stored := make(map[string]StoredCustomClaim)
	if record.Claims.Valid && record.Claims.String != "" {
		if err := json.Unmarshal([]byte(record.Claims.String), &stored); err != nil {
			return nil, false, fmt.Errorf("could not unmarshal existing custom claims: %w", err)
		}
	}

	wrote, valueChanged := mergeCustomClaims(stored, resolvedValues, managed, source)
	if !wrote {
		return nil, false, nil
	}

	claimsJSON, err := json.Marshal(stored)
	if err != nil {
		return nil, false, fmt.Errorf("could not marshal custom claims: %w", err)
	}
	record.Claims = nulls.NewString(string(claimsJSON))

	if err := persister.Update(record); err != nil {
		return nil, false, fmt.Errorf("could not update user custom claims: %w", err)
	}

	return record, valueChanged, nil
}
