package services

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/mail"
	"gopkg.in/gomail.v2"
)

type Email struct {
	renderer *mail.Renderer
	mailer   mail.Mailer
	cfg      config.Config
}

func NewEmailService(cfg config.Config) (*Email, error) {
	renderer, err := mail.NewRenderer()
	if err != nil {
		return nil, err
	}
	mailer, err := mail.NewMailer(cfg.Passcode.Smtp)
	if err != nil {
		panic(fmt.Errorf("failed to create mailer: %w", err))
	}

	return &Email{
		renderer,
		mailer,
		cfg,
	}, nil
}

// SendEmail sends an email with a translated specified template as body.
// The template name must be the name of the template without the content type and the file ending.
// E.g. when the file is created as "email_verification_text.tmpl" then the template name is just "email_verification"
// Currently only "[template_name]_text.tmpl" template can be used.
// The subject header of an email is also translated. The message_key must be "subject_[template_name]".
func (s *Email) SendEmail(template string, lang string, data map[string]interface{}, emailAddress string) error {
	text, err := s.renderer.Render(fmt.Sprintf("%s_text.tmpl", template), lang, data)
	if err != nil {
		return err
	}
	//html, err := s.renderer.Render(fmt.Sprintf("%s_html.tmpl", template), lang, data)
	if err != nil {
		return err
	}

	message := gomail.NewMessage()
	message.SetAddressHeader("To", emailAddress, "")
	message.SetAddressHeader("From", s.cfg.Passcode.Email.FromAddress, s.cfg.Passcode.Email.FromName)

	message.SetHeader("Subject", s.renderer.Translate(lang, fmt.Sprintf("subject_%s", template), data))
	message.SetBody("text/plain", text)
	//message.AddAlternative("text/html", html)

	err = s.mailer.Send(message)
	if err != nil {
		return err
	}

	return nil
}
