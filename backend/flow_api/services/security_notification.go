package services

import (
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	auditlog "github.com/teamhanko/hanko/backend/v2/audit_log"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"

	"github.com/teamhanko/hanko/backend/v2/config"
)

type SendSecurityNotificationParams struct {
	Template     string
	UserID       uuid.UUID
	EmailAddress string
	BodyData     map[string]interface{}
	HttpContext  echo.Context
	UserContext  models.User
}

type SendSecurityNotificationResult struct {
	SecurityNotificationModel models.SecurityNotification
	Subject                   string
	BodyPlain                 string
	BodyHTML                  string
}

type SecurityNotification interface {
	SendNotification(*pop.Connection, SendSecurityNotificationParams) error
}

type securityNotification struct {
	cfg          config.Config
	emailService Email
	auditLog     auditlog.Logger
	persister    persistence.Persister
}

func NewSecurityNotificationService(cfg config.Config, emailService Email, persister persistence.Persister, auditLog auditlog.Logger) SecurityNotification {
	return &securityNotification{
		cfg:          cfg,
		emailService: emailService,
		auditLog:     auditLog,
		persister:    persister,
	}
}

func (s securityNotification) SendNotification(tx *pop.Connection, p SendSecurityNotificationParams) error {
	language := p.HttpContext.Request().Header.Get("X-Language")

	subject := s.emailService.RenderSubject(language, p.Template, map[string]interface{}{
		"ServiceName": s.cfg.Service.Name,
	})

	if p.BodyData == nil {
		p.BodyData = map[string]interface{}{}
	}

	p.BodyData["ServiceName"] = s.cfg.Service.Name

	bodyPlain, err := s.emailService.RenderBodyPlain(language, p.Template, p.BodyData)
	if err != nil {
		return err
	}

	bodyHTML, err := s.emailService.RenderBodyHTML(language, p.Template, p.BodyData)
	if err != nil {
		return err
	}

	if s.cfg.EmailDelivery.Enabled {
		err = s.emailService.SendEmail(p.EmailAddress, subject, bodyPlain, bodyHTML)
		if err != nil {
			return err
		}

		auditLogDetails := []auditlog.DetailOption{
			auditlog.Detail("template", p.Template),
			auditlog.Detail("email_address", p.EmailAddress),
		}

		err := s.auditLog.CreateWithConnection(
			tx,
			p.HttpContext,
			models.AuditLogSecurityNotificationSent,
			&models.User{ID: p.UserID},
			nil,
			auditLogDetails...)

		if err != nil {
			return fmt.Errorf("could not create audit log: %w", err)
		}
	}

	return nil
}
