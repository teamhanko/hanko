package config

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
	"github.com/teamhanko/hanko/backend/v2/webhooks/validation"
)

type WebhookSecurityMode string

const (
	WebhookSecurityModePublicOnly   WebhookSecurityMode = "public_only"
	WebhookSecurityModeInternalOnly WebhookSecurityMode = "internal_only"
	WebhookSecurityModeCustom       WebhookSecurityMode = "custom"
	WebhookSecurityModeInsecure     WebhookSecurityMode = "insecure"
)

type WebhookSecurity struct {
	// `mode` defines the outbound destination policy for webhook callbacks.
	Mode WebhookSecurityMode `yaml:"mode" json:"mode,omitempty" koanf:"mode" jsonschema:"default=public_only,enum=public_only,enum=internal_only,enum=custom,enum=insecure"`
	// `allowed_schemes` defines the allowed URL schemes for webhook callbacks.
	AllowedSchemes []string `yaml:"allowed_schemes" json:"allowed_schemes,omitempty" koanf:"allowed_schemes"`
	// `follow_redirects` determines whether webhook delivery follows redirects.
	FollowRedirects bool `yaml:"follow_redirects" json:"follow_redirects,omitempty" koanf:"follow_redirects" jsonschema:"default=false"`
	// `max_redirects` defines the maximum number of redirects to follow.
	MaxRedirects int `yaml:"max_redirects" json:"max_redirects,omitempty" koanf:"max_redirects" jsonschema:"default=0"`
	// `skip_resolved_ip_validation` determines whether IP validation is skipped for hostnames
	// that passed hostname validation in custom mode. When true, if a hostname is in allowed_hosts
	// or allowed_domains, its resolved IPs are automatically trusted.
	// WARNING: Only enable this if you fully trust your DNS infrastructure. If DNS is compromised,
	// an attacker could make an allowed hostname resolve to a malicious IP (e.g., internal services).
	// When false (default), both hostname AND resolved IPs must be explicitly allowed (defense-in-depth).
	SkipResolvedIPValidation bool `yaml:"skip_resolved_ip_validation" json:"skip_resolved_ip_validation,omitempty" koanf:"skip_resolved_ip_validation" jsonschema:"default=false"`

	// `allowed_hosts` defines exact hostnames or IP addresses that are explicitly allowed in `custom` mode.
	// At least one of allowed_hosts, allowed_domains, or allowed_cidrs must be configured in custom mode.
	AllowedHosts []string `yaml:"allowed_hosts" json:"allowed_hosts,omitempty" koanf:"allowed_hosts"`
	// `allowed_domains` defines domains and subdomains that are explicitly allowed in `custom` mode.
	// At least one of allowed_hosts, allowed_domains, or allowed_cidrs must be configured in custom mode.
	AllowedDomains []string `yaml:"allowed_domains" json:"allowed_domains,omitempty" koanf:"allowed_domains"`
	// `allowed_cidrs` defines IP ranges that are explicitly allowed in `custom` mode.
	// At least one of allowed_hosts, allowed_domains, or allowed_cidrs must be configured in custom mode.
	AllowedCIDRs []string `yaml:"allowed_cidrs" json:"allowed_cidrs,omitempty" koanf:"allowed_cidrs"`

	// `blocked_hosts` defines exact hostnames that are blocked.
	BlockedHosts []string `yaml:"blocked_hosts" json:"blocked_hosts,omitempty" koanf:"blocked_hosts"`
	// `blocked_domains` defines domains and subdomains that are blocked.
	BlockedDomains []string `yaml:"blocked_domains" json:"blocked_domains,omitempty" koanf:"blocked_domains"`
	// `blocked_cidrs` defines IP ranges that are blocked.
	BlockedCIDRs []string `yaml:"blocked_cidrs" json:"blocked_cidrs,omitempty" koanf:"blocked_cidrs"`

	// `deny_metadata_endpoints` determines whether metadata service endpoints are blocked.
	DenyMetadataEndpoints bool `yaml:"deny_metadata_endpoints" json:"deny_metadata_endpoints,omitempty" koanf:"deny_metadata_endpoints" jsonschema:"default=true"`
	// `sanitize_errors` determines whether validation error messages are sanitized to prevent information disclosure.
	// When enabled, generic error messages are returned instead of detailed validation errors.
	// Detailed errors are still logged internally for debugging.
	SanitizeErrors bool `yaml:"sanitize_errors" json:"sanitize_errors,omitempty" koanf:"sanitize_errors" jsonschema:"default=false"`
}

// ToWebhookSecurityPolicy converts WebhookSecurity to validation.WebhookSecurityPolicy to avoid circular dependencies.
func (s *WebhookSecurity) ToWebhookSecurityPolicy() validation.WebhookSecurityPolicy {
	return validation.WebhookSecurityPolicy{
		Mode:                     validation.SecurityMode(s.Mode),
		AllowedSchemes:           s.AllowedSchemes,
		FollowRedirects:          s.FollowRedirects,
		MaxRedirects:             s.MaxRedirects,
		SkipResolvedIPValidation: s.SkipResolvedIPValidation,
		AllowedHosts:             s.AllowedHosts,
		AllowedDomains:           s.AllowedDomains,
		AllowedCIDRs:             s.AllowedCIDRs,
		BlockedHosts:             s.BlockedHosts,
		BlockedDomains:           s.BlockedDomains,
		BlockedCIDRs:             s.BlockedCIDRs,
		DenyMetadataEndpoints:    s.DenyMetadataEndpoints,
		SanitizeErrors:           s.SanitizeErrors,
	}
}

// GetAllowedSchemes returns the allowed URL schemes.
func (s *WebhookSecurity) GetAllowedSchemes() []string {
	return s.AllowedSchemes
}

func (s *WebhookSecurity) Validate() error {
	if err := s.validateMode(); err != nil {
		return err
	}

	if err := s.validateAllowedSchemes(); err != nil {
		return err
	}

	if err := s.validateRedirectSettings(); err != nil {
		return err
	}

	// Mode-specific validation
	if err := s.validateModeSpecificSettings(); err != nil {
		return err
	}

	// Only validate allowed/blocked lists if they will be used
	if s.Mode == WebhookSecurityModeCustom {
		// Custom mode requires at least one allowlist to be configured
		if len(s.AllowedHosts) == 0 && len(s.AllowedDomains) == 0 && len(s.AllowedCIDRs) == 0 {
			return fmt.Errorf("webhooks.security: custom mode requires at least one allow list (allowed_hosts, allowed_domains, or allowed_cidrs) to be configured. If you want to allow all destinations, use 'insecure' mode instead")
		}

		// Validate mutual exclusivity
		if err := s.validateMutualExclusivity(); err != nil {
			return err
		}

		if err := s.validateHostList("webhooks.security.allowed_hosts", s.AllowedHosts); err != nil {
			return err
		}

		if err := s.validateDomainList("webhooks.security.allowed_domains", s.AllowedDomains); err != nil {
			return err
		}

		if err := s.validateCIDRs("webhooks.security.allowed_cidrs", s.AllowedCIDRs); err != nil {
			return err
		}

		if err := s.validateHostList("webhooks.security.blocked_hosts", s.BlockedHosts); err != nil {
			return err
		}

		if err := s.validateDomainList("webhooks.security.blocked_domains", s.BlockedDomains); err != nil {
			return err
		}

		if err := s.validateCIDRs("webhooks.security.blocked_cidrs", s.BlockedCIDRs); err != nil {
			return err
		}
	}

	return nil
}

func (s *WebhookSecurity) validateMode() error {
	switch s.Mode {
	case WebhookSecurityModePublicOnly, WebhookSecurityModeInternalOnly, WebhookSecurityModeCustom, WebhookSecurityModeInsecure:
		return nil
	default:
		return fmt.Errorf("webhooks.security.mode must be one of: public_only, internal_only, custom, insecure")
	}
}

func (s *WebhookSecurity) validateAllowedSchemes() error {
	for i, scheme := range s.AllowedSchemes {
		switch strings.ToLower(strings.TrimSpace(scheme)) {
		case "http", "https":
		default:
			return fmt.Errorf("webhooks.security.allowed_schemes[%d] must be either 'http' or 'https'", i)
		}
	}

	return nil
}

func (s *WebhookSecurity) validateRedirectSettings() error {
	if !s.FollowRedirects && s.MaxRedirects != 0 {
		return fmt.Errorf("webhooks.security.max_redirects must be 0 when follow_redirects is false")
	}

	if s.MaxRedirects < 0 {
		return fmt.Errorf("webhooks.security.max_redirects must be greater than or equal to 0")
	}

	return nil
}

func (s *WebhookSecurity) validateCIDRs(field string, cidrs []string) error {
	for i, cidr := range cidrs {
		if _, _, err := net.ParseCIDR(strings.TrimSpace(cidr)); err != nil {
			return fmt.Errorf("%s[%d] is not a valid CIDR: %w", field, i, err)
		}
	}

	return nil
}

func (s *WebhookSecurity) validateHostList(field string, hosts []string) error {
	for i, host := range hosts {
		if validation.NormalizeHost(host) == "" {
			return fmt.Errorf("%s[%d] must not be empty", field, i)
		}
	}

	return nil
}

func (s *WebhookSecurity) validateDomainList(field string, domains []string) error {
	for i, domain := range domains {
		normalized := validation.NormalizeHost(domain)
		if normalized == "" {
			return fmt.Errorf("%s[%d] must not be empty", field, i)
		}
		if strings.Contains(normalized, ":") {
			return fmt.Errorf("%s[%d] must not contain a port", field, i)
		}
	}

	return nil
}

func (s *WebhookSecurity) validateModeSpecificSettings() error {
	// For public_only and internal_only modes, warn if allowed/blocked options are set
	if s.Mode == WebhookSecurityModePublicOnly || s.Mode == WebhookSecurityModeInternalOnly {
		var warnings []string

		if len(s.AllowedHosts) > 0 {
			warnings = append(warnings, "allowed_hosts")
		}
		if len(s.AllowedDomains) > 0 {
			warnings = append(warnings, "allowed_domains")
		}
		if len(s.AllowedCIDRs) > 0 {
			warnings = append(warnings, "allowed_cidrs")
		}
		if len(s.BlockedHosts) > 0 {
			warnings = append(warnings, "blocked_hosts")
		}
		if len(s.BlockedDomains) > 0 {
			warnings = append(warnings, "blocked_domains")
		}
		if len(s.BlockedCIDRs) > 0 {
			warnings = append(warnings, "blocked_cidrs")
		}

		if len(warnings) > 0 {
			log.Warn().
				Msgf("webhooks.security.mode=%s: the following set configuration options are ignored: %s (only allowed_schemes is effective in this mode)", s.Mode, strings.Join(warnings, ", "))
		}
	}

	return nil
}

func (s *WebhookSecurity) validateMutualExclusivity() error {
	// Check pairwise mutual exclusivity for custom mode
	if len(s.AllowedCIDRs) > 0 && len(s.BlockedCIDRs) > 0 {
		return fmt.Errorf("webhooks.security: allowed_cidrs and blocked_cidrs are mutually exclusive - use either allowlist or blocklist, not both")
	}

	if len(s.AllowedHosts) > 0 && len(s.BlockedHosts) > 0 {
		return fmt.Errorf("webhooks.security: allowed_hosts and blocked_hosts are mutually exclusive - use either allowlist or blocklist, not both")
	}

	if len(s.AllowedDomains) > 0 && len(s.BlockedDomains) > 0 {
		return fmt.Errorf("webhooks.security: allowed_domains and blocked_domains are mutually exclusive - use either allowlist or blocklist, not both")
	}

	return nil
}

type WebhookSettings struct {
	// `allow_time_expiration` determines whether webhooks are disabled when unused for 30 days
	// (only for database webhooks).
	AllowTimeExpiration bool `yaml:"allow_time_expiration" json:"allow_time_expiration,omitempty" koanf:"allow_time_expiration" jsonschema:"default=false"`
	// `enabled` enables the webhook feature.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	// `security` defines the outbound destination policy for webhook callbacks.
	Security WebhookSecurity `yaml:"security,omitempty" json:"security,omitempty" koanf:"security" jsonschema:"title=security"`
	// `hooks` is a list of Webhook configurations.
	//
	// When using environment variables the value for the `WEBHOOKS_HOOKS` key must be specified in the following
	// format:
	// `{"callback":"http://app.com/usercb","events":["user"]};{"callback":"http://app.com/emailcb","events":["email.send"]}`
	Hooks Webhooks `yaml:"hooks" json:"hooks,omitempty" koanf:"hooks" jsonschema:"title=hooks"`
}

func (ws *WebhookSettings) Validate() error {
	if !ws.Enabled {
		return nil
	}

	if err := ws.Security.Validate(); err != nil {
		return err
	}

	for i, hook := range ws.Hooks {
		if err := hook.Validate(&ws.Security); err != nil {
			return fmt.Errorf("webhooks.hooks[%d]: %w", i, err)
		}
	}

	return nil
}

type Webhooks []Webhook

// Decode is an implementation of the envconfig.Decoder interface.
// Assumes that environment variables (for the WEBHOOKS_HOOKS key) have the following format:
// {"callback":"http://app.com/usercb","events":["user"]};{"callback":"http://app.com/emailcb","events":["email.send"]}
func (wd *Webhooks) Decode(value string) error {
	webhooks := Webhooks{}
	hooks := strings.Split(value, ";")
	for _, hook := range hooks {
		webhook := Webhook{}
		err := json.Unmarshal([]byte(hook), &webhook)
		if err != nil {
			return fmt.Errorf("invalid map json: %w", err)
		}
		webhooks = append(webhooks, webhook)
	}

	*wd = webhooks
	return nil
}

type Webhook struct {
	// `callback` specifies the URL to which the change data will be sent.
	Callback string `yaml:"callback" json:"callback,omitempty" koanf:"callback"`
	// `events` is a list of events this hook listens for.
	Events events.Events `yaml:"events" json:"events,omitempty" koanf:"events" jsonschema:"title=events"`
}

func (Webhook) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Title = "hooks"
	evts, _ := schema.Properties.Get("events")

	// If the jsonschema.Reflector is configured with the DoNotReference option set to true, then the items property
	// in the schema is nil, hence we simply create a jsonschema.Schema manually, otherwise we'd get a nil pointer
	// exception.
	if evts.Items == nil {
		evts.Items = &jsonschema.Schema{Type: "string"}
	}
	evts.Items.Title = "events"
	evts.Items.Enum = []any{
		"user",
		"user.create",
		"user.delete",
		"user.login",
		"user.update",
		"user.update.email",
		"user.update.email.create",
		"user.update.email.delete",
		"user.update.email.primary",
		"user.update.password.update",
		"user.update.username",
		"user.update.username.create",
		"user.update.username.delete",
		"user.update.username.update",
		"email.send",
	}
	evts.Items.Extras = map[string]any{"meta:enum": map[string]string{
		"user":                        "Triggers on: user creation, user deletion, user update, email creation, email deletion, change of primary email",
		"user.create":                 "Triggers on: user creation",
		"user.delete":                 "Triggers on: user deletion",
		"user.login":                  "Triggers on: user login",
		"user.update":                 "Triggers on: user update, email creation, email deletion, change of primary email",
		"user.update.email":           "Triggers on: email creation, email deletion, change of primary email",
		"user.update.email.create":    "Triggers on: email creation",
		"user.update.email.delete":    "Triggers on: email deletion",
		"user.update.email.primary":   "Triggers on: change of primary email",
		"user.update.password.update": "Triggers on: change of password",
		"user.update.username":        "Triggers on: username creation, username deletion, change of username",
		"user.update.username.create": "Triggers on: username creation",
		"user.update.username.delete": "Triggers on: username deletion",
		"user.update.username.update": "Triggers on: change of username",
		"email.send":                  "Triggers on: an email was sent or should be sent",
	}}
}

func (w *Webhook) Validate(sec *WebhookSecurity) error {
	// Convert events.Events to []string for the shared validation
	eventsStr := make([]string, len(w.Events))
	for i, e := range w.Events {
		eventsStr[i] = string(e)
	}

	return validation.ValidateWebhook(w.Callback, eventsStr, sec)
}
