package shared

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sethvargo/go-limiter"
	auditlog "github.com/teamhanko/hanko/backend/v3/audit_log"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/ee/saml"
	"github.com/teamhanko/hanko/backend/v3/flow_api/services"
	"github.com/teamhanko/hanko/backend/v3/flowpilot"
	"github.com/teamhanko/hanko/backend/v3/mapper"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/session"
)

type Dependencies struct {
	Cfg                         config.Config
	HttpContext                 echo.Context
	SecurityNotificationService services.SecurityNotification
	PasscodeService             services.Passcode
	PasswordService             services.Password
	WebauthnService             services.WebauthnService
	SamlService                 saml.SamlProviderService
	Persister                   persistence.Persister
	SessionManager              session.Manager
	OTPRateLimiter              limiter.Store
	PasscodeRateLimiter         limiter.Store
	PasswordRateLimiter         limiter.Store
	TokenExchangeRateLimiter    limiter.Store
	Tx                          *pop.Connection
	AuthenticatorMetadata       mapper.AuthenticatorMetadata
	AuditLogger                 auditlog.Logger
	TenantID                    uuid.UUID
}

type Action struct{}

func (a *Action) GetDeps(c flowpilot.Context) *Dependencies {
	return c.Get("deps").(*Dependencies)
}
