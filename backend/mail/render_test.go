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

func TestRenderer_Render(t *testing.T) {
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
			Template: "loginTextMail",
			Lang:     "en",
			Expected: "Enter the following passcode on your login screen:\n\n123456\n\nThe passcode is valid for 5 minutes.",
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
			Template: "loginTextMail",
			Lang:     "xxx",
			Expected: "Enter the following passcode on your login screen:\n\n123456\n\nThe passcode is valid for 5 minutes.",
			WantErr:  false,
		},
		{
			Name:     "Login text template without translations for language",
			Template: "loginTextMail",
			Lang:     "es",
			Expected: "Enter the following passcode on your login screen:\n\n123456\n\nThe passcode is valid for 5 minutes.",
			WantErr:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result, err := renderer.Render(test.Template, test.Lang, templateData)

			if test.WantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.Expected, result)
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
			Expected: "Use passcode 123456 to sign in to Test Service",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := renderer.Translate(test.Lang, test.MessageID, test.Data)
			assert.Equal(t, test.Expected, result)
		})
	}
}
