package services

import (
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	auditlog "github.com/teamhanko/hanko/backend/v2/audit_log"
	"github.com/teamhanko/hanko/backend/v2/dto/webhook"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
	webhookUtils "github.com/teamhanko/hanko/backend/v2/webhooks/utils"

	"github.com/teamhanko/hanko/backend/v2/config"
)

type SendSecurityNotificationParams struct {
	Template     string
	UserID       uuid.UUID
	EmailAddress string
	BodyData     map[string]interface{}            // Data used in templates
	Data         *webhook.SecurityNotificationData // Data used for (serialized) webhook 'data' payload
	HttpContext  echo.Context
	UserContext  models.User
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

	deliveredByHanko := false
	if s.cfg.EmailDelivery.Enabled {
		err = s.emailService.SendEmail(p.EmailAddress, subject, bodyPlain, bodyHTML)
		if err != nil {
			return err
		}
		deliveredByHanko = true
	}

	if p.Data == nil {
		p.Data = &webhook.SecurityNotificationData{}
	}

	p.Data.Template = p.Template
	p.Data.ServiceName = s.cfg.Service.Name

	webhookData := webhook.EmailSend{
		Subject:          subject,
		BodyPlain:        bodyPlain,
		Body:             bodyHTML,
		ToEmailAddress:   p.EmailAddress,
		DeliveredByHanko: deliveredByHanko,
		Language:         language,
		Type:             "security_notification",
		Data:             p.Data,
	}

	err = webhookUtils.TriggerWebhooks(p.HttpContext, tx, events.EmailSend, webhookData)
	if err != nil {
		return err
	}

	// Prefer full user context if available; fall back to a bare user by ID.
	userForAudit := &models.User{ID: p.UserID}
	if p.UserContext.ID != uuid.Nil {
		userForAudit = &p.UserContext
	}

	auditLogDetails := []auditlog.DetailOption{
		auditlog.Detail("template", p.Template),
		auditlog.Detail("email_address", p.EmailAddress),
		auditlog.Detail("delivered_by_hanko", fmt.Sprintf("%t", deliveredByHanko)),
	}

	err = s.auditLog.CreateWithConnection(
		tx,
		p.HttpContext,
		models.AuditLogSecurityNotificationSent,
		userForAudit,
		nil,
		auditLogDetails...,
	)
	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	return nil
}
