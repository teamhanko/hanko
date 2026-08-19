package config

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	CustomClaimTypeString     = "string"
	CustomClaimTypeNumber     = "number"
	CustomClaimTypeBoolean    = "boolean"
	CustomClaimTypeStringList = "string_list"
)

// maxCustomClaimDefinitions bounds the number of tenant-wide custom claim definitions,
// to keep session JWTs and the user_custom_claims row from growing unbounded.
const maxCustomClaimDefinitions = 50

// customClaimNamePattern keeps names safe for gjson path lookups (CustomClaimsJWT.Get,
// dto/custom_claims.go), which treat `.`, `|`, `#`, `@`, `*`, `?` as meaningful
// metacharacters. Case is not restricted - that's the admin's choice, not ours.
var customClaimNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}$`)

// reservedCustomClaimNames mirrors the claim keys documented as always-ignored in
// Session.JWTTemplate (config_session.go) - custom claims must not shadow them.
// Compared case-insensitively in Validate() below; the admin's own casing is never rewritten.
var reservedCustomClaimNames = map[string]bool{
	"sub":        true,
	"iat":        true,
	"exp":        true,
	"aud":        true,
	"iss":        true,
	"email":      true,
	"username":   true,
	"session_id": true,
}

// CustomClaims configures the tenant-wide set of custom claims that SAML/OIDC connections
// may map their own attributes/claims onto.
type CustomClaims struct {
	// `definitions` declares the tenant-wide set of custom claims that connections may map onto.
	Definitions CustomClaimDefinitions `yaml:"definitions" json:"definitions,omitempty" koanf:"definitions" jsonschema:"title=definitions"`
}

// CustomClaimDefinitions is keyed by the custom claim name.
type CustomClaimDefinitions map[string]CustomClaimDefinition

type CustomClaimDefinition struct {
	// `name` is copied from the map key in TenantConfig.PostProcess, not configured directly.
	// Copied verbatim, not normalized - an admin typing an invalid (e.g. uppercase) name gets
	// a clear rejection from Validate() below rather than a silent rewrite.
	Name string `yaml:"-" json:"-" koanf:"-" jsonschema:"-"`
	// `type` declares how a mapped connection attribute value is coerced before storage.
	Type string `yaml:"type" json:"type" koanf:"type" jsonschema:"default=string,enum=string,enum=number,enum=boolean,enum=string_list"`
	// `description` is a human-readable note about the claim's meaning, for admin UIs/docs.
	Description string `yaml:"description" json:"description,omitempty" koanf:"description"`
}

func (d *CustomClaimDefinition) Validate() error {
	if !customClaimNamePattern.MatchString(d.Name) {
		return fmt.Errorf("name %q must match %s", d.Name, customClaimNamePattern.String())
	}
	if reservedCustomClaimNames[strings.ToLower(d.Name)] {
		return fmt.Errorf("name %q is reserved", d.Name)
	}

	switch d.Type {
	case CustomClaimTypeString, CustomClaimTypeNumber, CustomClaimTypeBoolean, CustomClaimTypeStringList:
	default:
		return fmt.Errorf("type %q must be one of %q, %q, %q, %q",
			d.Type, CustomClaimTypeString, CustomClaimTypeNumber, CustomClaimTypeBoolean, CustomClaimTypeStringList)
	}

	return nil
}

// PostProcess copies each map key into CustomClaimDefinition.Name, verbatim - no case
// normalization. An admin's chosen casing is preserved; Validate() below is what rejects
// a name that isn't safe for gjson path lookups, not this step.
func (c *CustomClaims) PostProcess() error {
	for key, definition := range c.Definitions {
		definition.Name = key
		c.Definitions[key] = definition
	}

	return nil
}

func (c *CustomClaims) Validate() error {
	if len(c.Definitions) > maxCustomClaimDefinitions {
		return fmt.Errorf("must not declare more than %d custom claim definitions", maxCustomClaimDefinitions)
	}

	for _, definition := range c.Definitions {
		if err := definition.Validate(); err != nil {
			return fmt.Errorf("failed to validate custom claim definition %s: %w", definition.Name, err)
		}
	}

	return nil
}
