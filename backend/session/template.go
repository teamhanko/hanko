package session

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/dto"
)

// ClaimTemplateData holds the data available for template processing
type ClaimTemplateData struct {
	User *dto.UserJWT
}

// ProcessClaimTemplate processes a map of claims using the provided user data and sets them on the token
func ProcessClaimTemplate(token jwt.Token, claims map[string]interface{}, user dto.UserJWT) error {
	claimTemplateData := ClaimTemplateData{
		User: &user,
	}
	for key, value := range claims {
		processedValue, err := processClaimValue(value, claimTemplateData)
		if err != nil {
			return fmt.Errorf("failed to process claim %s: %w", key, err)
		}
		_ = token.Set(key, processedValue)
	}
	return nil
}

// parseClaimTemplate parses and executes a template string using the provided data
func parseClaimTemplate(tmplStr string, data ClaimTemplateData) (string, error) {
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// processClaimValue processes a claim value, handling both string templates and nested structures
func processClaimValue(value interface{}, data ClaimTemplateData) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return parseClaimTemplate(v, data)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			processed, err := processClaimValue(val, data)
			if err != nil {
				return nil, err
			}
			result[key] = processed
		}
		return result, nil
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			processed, err := processClaimValue(val, data)
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
