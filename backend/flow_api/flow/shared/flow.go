package shared

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	"github.com/sethvargo/go-limiter"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/ee/saml"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/mapper"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/session"
)

type Dependencies struct {
	Cfg                   config.Config
	HttpContext           echo.Context
	PasscodeService       services.Passcode
	PasswordService       services.Password
	WebauthnService       services.WebauthnService
	SamlService           saml.Service
	Persister             persistence.Persister
	SessionManager        session.Manager
	RateLimiter           limiter.Store
	Tx                    *pop.Connection
	AuthenticatorMetadata mapper.AuthenticatorMetadata
	AuditLogger           auditlog.Logger
}

type Action struct{}

func (a *Action) GetDeps(c flowpilot.Context) *Dependencies {
	return c.Get("deps").(*Dependencies)
}
