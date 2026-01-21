package shared

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sethvargo/go-limiter"
	auditlog "github.com/teamhanko/hanko/backend/v2/audit_log"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/ee/saml"
	"github.com/teamhanko/hanko/backend/v2/flow_api/services"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
	"github.com/teamhanko/hanko/backend/v2/mapper"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/session"
)

type Dependencies struct {
	Cfg                         config.Config
	HttpContext                 echo.Context
	SecurityNotificationService services.SecurityNotification
	PasscodeService             services.Passcode
	PasswordService             services.Password
	WebauthnService             services.WebauthnService
	SamlService                 saml.Service
	Persister                   persistence.Persister
	SessionManager              session.Manager
	OTPRateLimiter              limiter.Store
	PasscodeRateLimiter         limiter.Store
	PasswordRateLimiter         limiter.Store
	TokenExchangeRateLimiter    limiter.Store
	Tx                          *pop.Connection
	AuthenticatorMetadata       mapper.AuthenticatorMetadata
	AuditLogger                 auditlog.Logger
	TenantID                    *uuid.UUID     // Tenant ID for multi-tenant mode
	Tenant                      *models.Tenant // Full tenant model for multi-tenant mode
}

type Action struct{}

func (a *Action) GetDeps(c flowpilot.Context) *Dependencies {
	return c.Get("deps").(*Dependencies)
}
