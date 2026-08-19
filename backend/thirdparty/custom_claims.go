package thirdparty

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	zeroLogger "github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence"
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
// CustomThirdPartyProvider isn't necessarily OIDC-conformant), or "admin" for a value set via
// the Admin API. Tracked so applyCustomClaims only ever lets the connection that set a claim
// clear it again - see applyCustomClaims's doc comment for why that matters.
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
// connection under its new identity, so it can no longer clear them itself (an admin can
// still fix this via the Admin API PATCH).
func customClaimConnectionSource(providerID string, isSaml bool) string {
	if isSaml {
		return "saml:" + providerID
	}
	return "third_party:" + providerID
}

// applyCustomClaims resolves userData.CustomClaimSource against the tenant's declared
// definitions and merges the result into the user's persisted custom claims. Returns whether
// anything actually changed, so callers can decide whether to surface a webhook event.
//
// Refresh semantics: a claim this connection manages is always overwritten with its freshly
// resolved value - last-one-wins, any source, same as the write half of
// User.SyncFromProviderProfile. A claim this connection manages but found no value for is
// cleared only if the currently stored value's source is this same connection: the one
// connection that owns a claim losing its own assertion is a real signal (e.g. a student
// graduating), but an unrelated connection's momentary silence about a claim it never set
// must never wipe it - only the connection that actually set a value gets to take it away.
//
// Unlike ResolveCustomClaims (which never fails - bad claim *data* is tolerated by design),
// persistence errors here propagate, consistent with how this file already treats every
// other write in the same transaction (e.g. User.SyncFromProviderProfile's Update below).
func applyCustomClaims(tx *pop.Connection, p persistence.Persister, cfg *config.TenantConfig, userData *UserData, source string, userID uuid.UUID, tenantID uuid.UUID) (changed bool, err error) {
	if userData.CustomClaimSource == nil {
		return false, nil
	}

	resolvedValues, managed := ResolveCustomClaims(cfg.CustomClaims.Definitions, userData.CustomClaimSource)
	if len(managed) == 0 {
		return false, nil
	}

	persister := p.GetUserCustomClaimsPersisterWithConnection(tx)
	record, err := persister.Get(userID, tenantID)
	if err != nil {
		return false, fmt.Errorf("could not get user custom claims: %w", err)
	}

	stored := make(map[string]StoredCustomClaim)
	if len(record.Claims) > 0 {
		if err := json.Unmarshal(record.Claims, &stored); err != nil {
			return false, fmt.Errorf("could not unmarshal existing custom claims: %w", err)
		}
	}

	for _, claimName := range managed {
		if value, ok := resolvedValues[claimName]; ok {
			stored[claimName] = StoredCustomClaim{Value: value, Source: source}
			changed = true
			continue
		}

		if existing, exists := stored[claimName]; exists && existing.Source == source {
			delete(stored, claimName)
			changed = true
		}
	}

	if !changed {
		return false, nil
	}

	claimsJSON, err := json.Marshal(stored)
	if err != nil {
		return false, fmt.Errorf("could not marshal custom claims: %w", err)
	}
	record.Claims = claimsJSON

	if err := persister.Update(record); err != nil {
		return false, fmt.Errorf("could not update user custom claims: %w", err)
	}

	return true, nil
}
