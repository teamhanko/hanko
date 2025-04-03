package session

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/dto"
)

// ClaimTemplateData holds the data available for template processing
type ClaimTemplateData struct {
	User *dto.UserJWT
}

// ProcessJWTTemplate processes a map of claims using the provided user data and sets them on the token
func ProcessJWTTemplate(token jwt.Token, claims map[string]interface{}, user dto.UserJWT) error {
	claimTemplateData := ClaimTemplateData{
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
func processClaimTemplate(value interface{}, data ClaimTemplateData) (interface{}, error) {
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
func parseClaimTemplateValue(tmplStr string, data ClaimTemplateData) (interface{}, error) {
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// "Workaround"/"hack" for when the template expression evaluates to a boolean string, i.e. "true"
	// or "false". This converts it to a bool for consistency's sake (i.e. to prevent that both boolean
	// values and boolean strings are eventually set in the JWT).
	resultString := buf.String()
	if resultString == "true" || resultString == "false" {
		b, err := strconv.ParseBool(buf.String())
		if err != nil {
			return nil, fmt.Errorf("could not parse string as bool: %w", err)
		}

		return b, nil
	}

	return resultString, nil
}
