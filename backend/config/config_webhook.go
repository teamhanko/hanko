package config

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
)

type WebhookSecurityMode string

const (
	WebhookSecurityModePublicOnly WebhookSecurityMode = "public_only"
	WebhookSecurityModeCustom     WebhookSecurityMode = "custom"
	WebhookSecurityModeInsecure   WebhookSecurityMode = "insecure"
)

type WebhookSecurity struct {
	// `mode` defines the outbound destination policy for webhook callbacks.
	Mode WebhookSecurityMode `yaml:"mode" json:"mode,omitempty" koanf:"mode" jsonschema:"default=public_only,enum=public_only,enum=custom,enum=insecure"`
	// `allowed_schemes` defines the allowed URL schemes for webhook callbacks.
	AllowedSchemes []string `yaml:"allowed_schemes" json:"allowed_schemes,omitempty" koanf:"allowed_schemes"`
	// `follow_redirects` determines whether webhook delivery follows redirects.
	FollowRedirects bool `yaml:"follow_redirects" json:"follow_redirects,omitempty" koanf:"follow_redirects" jsonschema:"default=false"`
	// `max_redirects` defines the maximum number of redirects to follow.
	MaxRedirects int `yaml:"max_redirects" json:"max_redirects,omitempty" koanf:"max_redirects" jsonschema:"default=0"`

	// `allowed_hosts` defines exact hostnames that are explicitly allowed in `custom` mode.
	AllowedHosts []string `yaml:"allowed_hosts" json:"allowed_hosts,omitempty" koanf:"allowed_hosts"`
	// `allowed_domains` defines domains and subdomains that are explicitly allowed in `custom` mode.
	AllowedDomains []string `yaml:"allowed_domains" json:"allowed_domains,omitempty" koanf:"allowed_domains"`
	// `allowed_cidrs` defines IP ranges that are explicitly allowed in `custom` mode.
	AllowedCIDRs []string `yaml:"allowed_cidrs" json:"allowed_cidrs,omitempty" koanf:"allowed_cidrs"`

	// `blocked_hosts` defines exact hostnames that are blocked.
	BlockedHosts []string `yaml:"blocked_hosts" json:"blocked_hosts,omitempty" koanf:"blocked_hosts"`
	// `blocked_domains` defines domains and subdomains that are blocked.
	BlockedDomains []string `yaml:"blocked_domains" json:"blocked_domains,omitempty" koanf:"blocked_domains"`
	// `blocked_cidrs` defines IP ranges that are blocked.
	BlockedCIDRs []string `yaml:"blocked_cidrs" json:"blocked_cidrs,omitempty" koanf:"blocked_cidrs"`

	// `deny_metadata_endpoints` determines whether metadata service endpoints are blocked.
	DenyMetadataEndpoints bool `yaml:"deny_metadata_endpoints" json:"deny_metadata_endpoints,omitempty" koanf:"deny_metadata_endpoints" jsonschema:"default=true"`
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

	return nil
}

func (s *WebhookSecurity) validateMode() error {
	switch s.Mode {
	case WebhookSecurityModePublicOnly, WebhookSecurityModeCustom, WebhookSecurityModeInsecure:
		return nil
	default:
		return fmt.Errorf("webhooks.security.mode must be one of: public_only, custom, insecure")
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
		if normalizeWebhookHost(host) == "" {
			return fmt.Errorf("%s[%d] must not be empty", field, i)
		}
	}

	return nil
}

func (s *WebhookSecurity) validateDomainList(field string, domains []string) error {
	for i, domain := range domains {
		normalized := normalizeWebhookHost(domain)
		if normalized == "" {
			return fmt.Errorf("%s[%d] must not be empty", field, i)
		}
		if strings.Contains(normalized, ":") {
			return fmt.Errorf("%s[%d] must not contain a port", field, i)
		}
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
	Security WebhookSecurity `yaml:"security" json:"security,omitempty" koanf:"security" jsonschema:"title=security"`
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
	parsed, err := url.Parse(w.Callback)
	if err != nil {
		return fmt.Errorf("callback is not a valid URL: %w", err)
	}

	if err := w.validateParsedURL(parsed, sec); err != nil {
		return err
	}

	if err := w.validateLiteralIP(parsed, sec); err != nil {
		return err
	}

	if err := w.validateEvents(); err != nil {
		return err
	}

	return nil
}

func (w *Webhook) validateParsedURL(parsed *url.URL, sec *WebhookSecurity) error {
	if parsed.Scheme == "" {
		return fmt.Errorf("callback URL must include a scheme")
	}

	if parsed.Host == "" {
		return fmt.Errorf("callback URL must include a host")
	}

	if parsed.User != nil {
		return fmt.Errorf("callback URL must not include user info")
	}

	schemeAllowed := false
	for _, scheme := range sec.AllowedSchemes {
		if strings.EqualFold(strings.TrimSpace(scheme), parsed.Scheme) {
			schemeAllowed = true
			break
		}
	}

	if !schemeAllowed {
		return fmt.Errorf("callback scheme '%s' is not allowed", parsed.Scheme)
	}

	host := normalizeWebhookHost(parsed.Hostname())
	if host == "" {
		return fmt.Errorf("callback host must not be empty")
	}

	if matchesHost(host, sec.BlockedHosts) {
		return fmt.Errorf("callback host '%s' is blocked", host)
	}

	if matchesDomain(host, sec.BlockedDomains) {
		return fmt.Errorf("callback host '%s' matches a blocked domain", host)
	}

	if err := w.validateAllowedHostPolicy(host, sec); err != nil {
		return err
	}

	return nil
}

func (w *Webhook) validateAllowedHostPolicy(host string, sec *WebhookSecurity) error {
	switch sec.Mode {
	case WebhookSecurityModeInsecure:
		return nil
	case WebhookSecurityModePublicOnly:
		return nil
	case WebhookSecurityModeCustom:
		if len(sec.AllowedHosts) == 0 && len(sec.AllowedDomains) == 0 {
			return nil
		}

		if matchesHost(host, sec.AllowedHosts) || matchesDomain(host, sec.AllowedDomains) {
			return nil
		}

		return fmt.Errorf("callback host '%s' is not in the allowed host/domain list", host)
	default:
		return fmt.Errorf("unsupported webhook security mode '%s'", sec.Mode)
	}
}

func (w *Webhook) validateLiteralIP(parsed *url.URL, sec *WebhookSecurity) error {
	host := normalizeWebhookHost(parsed.Hostname())
	ip := net.ParseIP(host)
	if ip == nil {
		return nil
	}

	if err := w.validateAbsoluteDenies(ip, sec); err != nil {
		return err
	}

	return w.validateModeDecision(ip, sec)
}

func (w *Webhook) validateAbsoluteDenies(ip net.IP, sec *WebhookSecurity) error {
	if ipMatchesCIDRs(ip, sec.BlockedCIDRs) {
		return fmt.Errorf("callback IP '%s' is blocked", ip.String())
	}

	if sec.DenyMetadataEndpoints && isMetadataIP(ip) {
		return fmt.Errorf("metadata endpoint IP '%s' is blocked", ip.String())
	}

	return nil
}

func (w *Webhook) validateModeDecision(ip net.IP, sec *WebhookSecurity) error {
	switch sec.Mode {
	case WebhookSecurityModeInsecure:
		return nil
	case WebhookSecurityModePublicOnly:
		return w.validatePublicOnly(ip)
	case WebhookSecurityModeCustom:
		return w.validateCustom(ip, sec)
	default:
		return fmt.Errorf("unsupported webhook security mode '%s'", sec.Mode)
	}
}

func (w *Webhook) validatePublicOnly(ip net.IP) error {
	if !isPublicRoutableIP(ip) {
		return fmt.Errorf("non-public IP '%s' is not allowed in public_only mode", ip.String())
	}

	return nil
}

func (w *Webhook) validateCustom(ip net.IP, sec *WebhookSecurity) error {
	if ipMatchesCIDRs(ip, sec.AllowedCIDRs) {
		return nil
	}

	if !isPublicRoutableIP(ip) {
		return fmt.Errorf("non-public IP '%s' is not allowed unless explicitly allowlisted", ip.String())
	}

	return nil
}

func (w *Webhook) validateEvents() error {
	if len(w.Events) > 0 {
		for i, e := range w.Events {
			isValid := events.IsValidEvent(e)
			if !isValid {
				return fmt.Errorf("event '%s' at events[%d] is not a valid webhook event", e, i)
			}
		}
	}

	return nil
}

func normalizeWebhookHost(value string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
}

func matchesHost(host string, values []string) bool {
	host = normalizeWebhookHost(host)
	for _, value := range values {
		if host == normalizeWebhookHost(value) {
			return true
		}
	}
	return false
}

func matchesDomain(host string, domains []string) bool {
	host = normalizeWebhookHost(host)
	for _, domain := range domains {
		normalized := normalizeWebhookHost(domain)
		if host == normalized || strings.HasSuffix(host, "."+normalized) {
			return true
		}
	}
	return false
}

func ipMatchesCIDRs(ip net.IP, cidrs []string) bool {
	for _, value := range cidrs {
		_, network, err := net.ParseCIDR(strings.TrimSpace(value))
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

func isPublicRoutableIP(ip net.IP) bool {
	if ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsMulticast() ||
		ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		isReservedIP(ip) ||
		isMetadataIP(ip) {
		return false
	}

	return true
}

func isReservedIP(ip net.IP) bool {
	return ipMatchesCIDRs(ip, []string{
		"0.0.0.0/8",
		"100.64.0.0/10",
		"192.0.0.0/24",
		"192.0.2.0/24",
		"198.18.0.0/15",
		"198.51.100.0/24",
		"203.0.113.0/24",
		"240.0.0.0/4",
		"::/128",
		"100::/64",
		"2001:db8::/32",
	})
}

func isMetadataIP(ip net.IP) bool {
	return ipMatchesCIDRs(ip, []string{
		"169.254.169.254/32",
	})
}
