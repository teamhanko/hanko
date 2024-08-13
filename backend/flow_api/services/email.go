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
	mailer, err := mail.NewMailer(cfg.EmailDelivery.SMTP)
	if err != nil {
		panic(fmt.Errorf("failed to create mailer: %w", err))
	}

	return &Email{
		renderer,
		mailer,
		cfg,
	}, nil
}

// SendEmail sends an email to the emailAddress with the given subject and body.
func (s *Email) SendEmail(emailAddress, subject, body string) error {
	message := gomail.NewMessage()
	message.SetAddressHeader("To", emailAddress, "")
	message.SetAddressHeader("From", s.cfg.EmailDelivery.FromAddress, s.cfg.EmailDelivery.FromName)
	message.SetHeader("Subject", subject)
	message.SetBody("text/plain", body)

	if err := s.mailer.Send(message); err != nil {
		return err
	}

	return nil
}

// RenderSubject renders a subject with the given template. Must be "subject_[template_name]".
func (s *Email) RenderSubject(lang, template string, data map[string]interface{}) string {
	return s.renderer.Translate(lang, fmt.Sprintf("subject_%s", template), data)
}

// RenderBody renders the body with the given template. The template name must be the name of the template without the
// content type and the file ending. E.g. when the file is created as "email_verification_text.tmpl" then the template
// name is just "email_verification"
func (s *Email) RenderBody(lang, template string, data map[string]interface{}) (string, error) {
	return s.renderer.Render(fmt.Sprintf("%s_text.tmpl", template), lang, data)
}
