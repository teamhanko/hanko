package session

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/teamhanko/hanko/backend/dto"
)

// TemplateData holds the data available for template processing
type TemplateData struct {
	User *dto.UserJWT
}

// processTemplate processes a template string using the provided data
func processTemplate(tmplStr string, data TemplateData) (string, error) {
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
func processClaimValue(value interface{}, data TemplateData) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return processTemplate(v, data)
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
