package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/dto"
)

// JWTTemplateData holds the data available for template processing
type JWTTemplateData struct {
	User *dto.UserJWT
}

// ProcessJWTTemplate processes a map of claims using the provided user data and sets them on the token
func ProcessJWTTemplate(token jwt.Token, claims map[string]interface{}, user dto.UserJWT) error {
	claimTemplateData := JWTTemplateData{
		User: &user,
	}
	for key, value := range claims {
		processedValue, err := processClaimTemplate(value, claimTemplateData)
		if err != nil {
			log.Warn().Err(err).Str("session", key).Msgf("failed to process custom JWT claim template: %+v", value)
			continue
		}
		err = token.Set(key, processedValue)
		if err != nil {
			log.Warn().Err(err).Str("session", key).Msgf("failed to set processed JWT claim %+v to token", value)
			continue
		}
	}
	return nil
}

// processClaimTemplate processes a claim value, handling both string templates and nested structures
func processClaimTemplate(value interface{}, data JWTTemplateData) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return parseClaimTemplateValue(v, data)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			processed, err := processClaimTemplate(val, data)
			if err != nil {
				return nil, err
			}
			result[key] = processed
		}
		return result, nil
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			processed, err := processClaimTemplate(val, data)
			if err != nil {
				return nil, err
			}
			result[i] = processed
		}
		return result, nil
	default:
		return value, nil
	}
}

// parseClaimTemplateValue parses and executes a template string using the provided data
func parseClaimTemplateValue(tmplStr string, data JWTTemplateData) (interface{}, error) {
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	resultString := buf.String()

	// "Workaround"/"hack" for when the template expression evaluates to a boolean string, i.e. "true"
	// or "false". This converts it to a bool for consistency's sake (i.e. to prevent that both boolean
	// values and boolean strings are eventually set in the JWT).
	if resultString == "true" || resultString == "false" {
		b, err := strconv.ParseBool(resultString)
		if err != nil {
			return nil, fmt.Errorf("could not parse string as bool: %w", err)
		}
		return b, nil
	}

	// Another workaround for JSON objects and arrays. Somewhere along the way, representations
	// of JSON objects and arrays end up with multiple escape characters. This was the only way
	// to properly get rid of them.
	looksLikeObject := strings.HasPrefix(resultString, "{") && strings.HasSuffix(resultString, "}")
	looksLikeArray := strings.HasPrefix(resultString, "[") && strings.HasSuffix(resultString, "]")
	if looksLikeObject || looksLikeArray {
		var result interface{}
		if err := json.Unmarshal([]byte(resultString), &result); err == nil {
			return result, nil
		}
	}

	return resultString, nil
}
