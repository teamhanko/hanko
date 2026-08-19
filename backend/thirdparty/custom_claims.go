package thirdparty

import (
	"strconv"

	zeroLogger "github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/v3/config"
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
