package mail

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/config"
	"gopkg.in/gomail.v2"
	"strconv"
)

type Mailer interface {
	Send(message *gomail.Message) error
}

type mailer struct {
	dialer *gomail.Dialer
}

func NewMailer(config config.SMTP) (Mailer, error) {
	port, err := strconv.Atoi(config.Port)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SMTP port: %w", err)
	}
	d := gomail.NewDialer(config.Host, port, config.User, config.Password)
	return &mailer{
		dialer: d,
	}, nil
}

func (m *mailer) Send(message *gomail.Message) error {
	return m.dialer.DialAndSend(message)
}
