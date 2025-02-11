package mail

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRenderer(t *testing.T) {
	renderer, err := NewRenderer()

	assert.NoError(t, err)
	assert.NotEmpty(t, renderer)
}

func TestRenderer_RenderPlain(t *testing.T) {
	renderer, err := NewRenderer()

	assert.NoError(t, err)
	assert.NotEmpty(t, renderer)

	templateData := map[string]interface{}{
		"TTL":  5,
		"Code": "123456",
	}

	tests := []struct {
		Name     string
		Template string
		Lang     string
		Expected string
		WantErr  bool
	}{
		{
			Name:     "Login text template",
			Template: "login",
			Lang:     "en",
			Expected: "Enter the following passcode to verify your identity:\n\n123456\n\nThe passcode is valid for 5 minutes.",
			WantErr:  false,
		},
		{
			Name:     "Not existing template",
			Template: "NotExistingTemplate",
			Lang:     "en",
			Expected: "",
			WantErr:  true,
		},
		{
			Name:     "Login text template with unknown language",
			Template: "login",
			Lang:     "xxx",
			Expected: "Enter the following passcode to verify your identity:\n\n123456\n\nThe passcode is valid for 5 minutes.",
			WantErr:  false,
		},
		{
			Name:     "Login text template without translations for language",
			Template: "login",
			Lang:     "es",
			Expected: "Enter the following passcode to verify your identity:\n\n123456\n\nThe passcode is valid for 5 minutes.",
			WantErr:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result, err := renderer.RenderPlain(test.Template, test.Lang, templateData)

			if test.WantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.Expected, result)
			}
		})
	}
}

func TestRenderer_RenderHTML(t *testing.T) {
	renderer, err := NewRenderer()

	assert.NoError(t, err)
	assert.NotEmpty(t, renderer)

	templateData := map[string]interface{}{
		"TTL":  5,
		"Code": "123456",
	}

	tests := []struct {
		Name     string
		Template string
		Lang     string
		Expected []string
		WantErr  bool
	}{
		{
			Name:     "Login text template",
			Template: "login",
			Lang:     "en",
			Expected: []string{"<!DOCTYPE html>", "Enter the following passcode to verify your identity:", "123456", "The passcode is valid for 5 minutes."},
			WantErr:  false,
		},
		{
			Name:     "Not existing template",
			Template: "NotExistingTemplate",
			Lang:     "en",
			Expected: []string{"<!DOCTYPE html>"},
			WantErr:  true,
		},
		{
			Name:     "Login text template with unknown language",
			Template: "login",
			Lang:     "xxx",
			Expected: []string{"<!DOCTYPE html>", "Enter the following passcode to verify your identity:", "123456", "The passcode is valid for 5 minutes."},
			WantErr:  false,
		},
		{
			Name:     "Login text template without translations for language",
			Template: "login",
			Lang:     "es",
			Expected: []string{"<!DOCTYPE html>", "Enter the following passcode to verify your identity:", "123456", "The passcode is valid for 5 minutes."},
			WantErr:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result, err := renderer.RenderHTML(test.Template, test.Lang, templateData)

			if test.WantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, expected := range test.Expected {
					assert.Contains(t, result, expected)
				}
			}
		})
	}
}

func TestRenderer_Translate(t *testing.T) {
	renderer, err := NewRenderer()

	assert.NoError(t, err)
	assert.NotEmpty(t, renderer)

	tests := []struct {
		Name      string
		MessageID string
		Lang      string
		Data      map[string]interface{}
		Expected  string
	}{
		{
			Name:      "Translate email_subject_login",
			MessageID: "email_subject_login",
			Lang:      "en",
			Data: map[string]interface{}{
				"ServiceName": "Test Service",
				"Code":        "123456",
			},
			Expected: "123456 is your passcode for Test Service",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := renderer.Translate(test.Lang, test.MessageID, test.Data)
			assert.Equal(t, test.Expected, result)
		})
	}
}
