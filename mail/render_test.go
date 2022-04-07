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
		"UserEmail":     "example@example.com",
		"ServiceDomain": "example.com",
		"Ttl":           5,
		"Code":          "123456",
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
			Expected: "\nEnter the following passcode in your login screen at example.com to sign in as example@example.com:\n\n123456\n\nThe passcode is valid for 5 minutes.\n",
			WantErr:  false,
		},
		{
			Name:     "Password Recovery template",
			Template: "passwordRecoveryTextMail",
			Lang:     "en",
			Expected: "\nA password reset for example@example.com was requested at example.com. Enter the following passcode to access your account and create a new password:\n\n123456\n\nThe passcode is valid for 5 minutes.\n",
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
			Expected: "\nEnter the following passcode in your login screen at example.com to sign in as example@example.com:\n\n123456\n\nThe passcode is valid for 5 minutes.\n",
			WantErr:  false,
		},
		{
			Name:     "Login text template without translations for language",
			Template: "loginTextMail",
			Lang:     "es",
			Expected: "\nEnter the following passcode in your login screen at example.com to sign in as example@example.com:\n\n123456\n\nThe passcode is valid for 5 minutes.\n",
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
