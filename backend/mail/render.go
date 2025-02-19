package mail

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
	"html/template"
	"strings"
)

//go:embed templates/* locales/*
var mailFS embed.FS

type Renderer struct {
	templatePlain *template.Template
	bundle        *i18n.Bundle
	localizer     *i18n.Localizer
}

// NewRenderer creates an instance of Renderer, which renders the templates (located in mail/templates) with locales (located in mail/locales)
func NewRenderer() (*Renderer, error) {
	r := &Renderer{}
	bundle := i18n.NewBundle(language.English)
	dir, err := mailFS.ReadDir("locales")
	if err != nil {
		return nil, fmt.Errorf("failed to read locales directory: %w", err)
	}
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	for _, entry := range dir {
		_, _ = bundle.LoadMessageFileFS(mailFS, fmt.Sprintf("locales/%s", entry.Name()))
	}
	r.bundle = bundle

	// add the translate function to the template, so it can be used inside
	funcMap := template.FuncMap{"t": r.translate}
	t := template.New("root").Funcs(funcMap)
	_, err = t.ParseFS(mailFS, "templates/*.txt.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}
	r.templatePlain = t

	return r, nil
}

// translate is a helper function to translate texts in a template
func (r *Renderer) translate(messageID string, templateData map[string]interface{}) string {
	localizer := i18n.NewLocalizer(r.bundle, templateData["renderer_lang"].(string))
	return localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
}

// RenderPlain renders a template with the given data and lang.
// The lang can be the contents of Accept-Language headers as defined in http://www.ietf.org/rfc/rfc2616.txt.
func (r *Renderer) RenderPlain(templateName string, lang string, data map[string]interface{}) (string, error) {
	r.localizer = i18n.NewLocalizer(r.bundle, lang) // set the localizer, so the test will be translated to the given language
	data["renderer_lang"] = lang
	templateBuffer := &bytes.Buffer{}
	err := r.templatePlain.ExecuteTemplate(templateBuffer, fmt.Sprintf("%s.txt.tmpl", templateName), data)
	if err != nil {
		return "", fmt.Errorf("failed to fill plain text template with data: %w", err)
	}
	return strings.TrimSpace(templateBuffer.String()), nil
}

// RenderHTML renders an HTML template with the given data and lang.
// The lang can be the contents of Accept-Language headers as defined in http://www.ietf.org/rfc/rfc2616.txt.
func (r *Renderer) RenderHTML(templateName string, lang string, data map[string]interface{}) (string, error) {
	var buffer bytes.Buffer

	r.localizer = i18n.NewLocalizer(r.bundle, lang)
	data["renderer_lang"] = lang

	templateHTML := template.New("root").Funcs(template.FuncMap{"t": r.translate})
	patterns := []string{"templates/layout.html.tmpl", fmt.Sprintf("templates/%s.html.tmpl", templateName)}
	_, err := templateHTML.ParseFS(mailFS, patterns...)
	if err != nil {
		return "", fmt.Errorf("failed to parse html template: %w", err)
	}

	err = templateHTML.ExecuteTemplate(&buffer, "layout", data)
	if err != nil {
		return "", fmt.Errorf("failed to execute html template: %w", err)
	}

	return strings.TrimSpace(buffer.String()), nil
}

func (r *Renderer) Translate(lang string, messageID string, data map[string]interface{}) string {
	loc := i18n.NewLocalizer(r.bundle, lang)
	return loc.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
}
