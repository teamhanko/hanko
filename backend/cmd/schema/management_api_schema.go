package schema

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/invopop/jsonschema"
	"github.com/teamhanko/hanko/backend/v3/config"
)

// managementAPITenantConfigDef is the name invopop gives config.TenantConfig when reflected
// directly (as opposed to via config.Config, where it's inlined/flattened into Config's own
// properties).
const managementAPITenantConfigDef = "TenantConfig"

// optionalTenantConfigRefs holds $defs names that must stay optional in the response schema even
// though they're declared properties, because the corresponding Go field is a pointer with a
// genuine "absent" meaning (nil), not merely a value with a usable zero-value default. Currently
// only Session.JWTTemplate.
var optionalTenantConfigRefs = map[string]bool{
	"JWTTemplate": true,
}

// needsCreateVariant holds $defs names that declare (or contain a descendant that declares) an
// if/then/else conditional rule, and therefore need their own *CreateInput schema variant — see
// buildManagementAPISchemas for why Create can safely publish these rules but Update can't.
var needsCreateVariant = map[string]bool{
	"Secrets":                   true,
	"KeyManagement":             true,
	"ThirdParty":                true,
	"ThirdPartyProviders":       true,
	"ThirdPartyProvider":        true,
	"CustomThirdPartyProviders": true,
	"CustomThirdPartyProvider":  true,
}

// handWrittenManagementAPISchemas are the schema names this generator never touches — anything not
// derived from config.TenantConfig by reflection (envelope/DTO/status types authored directly in
// the spec).
var handWrittenManagementAPISchemas = []string{
	"Error",
	"HealthAliveStatus",
	"HealthReadyStatus",
	"CreateTenantRequest",
	"UpdateTenantRequest",
	"Tenant",
	"CreateSamlProviderRequest",
	"UpdateSamlProviderRequest",
	"SamlProvider",
}

type generatedSchema struct {
	name   string
	schema *jsonschema.Schema
}

// buildManagementAPISchemas reflects config.TenantConfig and derives three variants of every
// schema it depends on:
//
//   - the plain name (e.g. "ThirdParty"): the response shape, as returned by this API. Every
//     property is required except optionalTenantConfigRefs, and if/then/else conditionals (e.g.
//     "enabled ⇒ client_id required") are kept, since both are true facts about a persisted,
//     fully-resolved resource.
//   - "*Input" (e.g. "ThirdPartyInput"): the request shape used by PUT /tenants/{id}. Nothing is
//     required and no conditionals apply, because Update merges the request onto the tenant's
//     current (per-tenant, unknowable-in-a-static-schema) config — a partial request might rely on
//     a value already stored from an earlier request that this schema can't see.
//   - "*CreateInput" (e.g. "ThirdPartyCreateInput"), only for needsCreateVariant: the request shape
//     used by POST /tenants. Still nothing unconditionally required (config.DefaultTenantConfig()
//     covers every field), but conditionals ARE kept, because Create's merge base is the static,
//     known default document — a conditional like "enabled ⇒ client_id required" is always either
//     true or false for a given request, independent of any other tenant's state.
func buildManagementAPISchemas() ([]generatedSchema, error) {
	reflector := &jsonschema.Reflector{}
	if err := reflector.AddGoComments("github.com/teamhanko/hanko/backend/v3", "config"); err != nil {
		return nil, fmt.Errorf("failed to add go comments: %w", err)
	}

	root := reflector.Reflect(&config.TenantConfig{})
	defs := root.Definitions
	if _, ok := defs[managementAPITenantConfigDef]; !ok {
		return nil, fmt.Errorf("reflecting config.TenantConfig did not produce a %q definition", managementAPITenantConfigDef)
	}

	response := make(map[string]*jsonschema.Schema, len(defs))
	for name, def := range defs {
		s := cloneSchema(def)
		rewriteSchemaRefs(s, identitySuffix)
		computeRequired(s)
		response[name] = s
	}

	input := make(map[string]*jsonschema.Schema, len(defs))
	for name, def := range response {
		s := cloneSchema(def)
		stripRequiredAndConditionals(s)
		rewriteSchemaRefs(s, suffixWith("Input"))
		input[name+"Input"] = s
	}

	createInputResolve := func(name string) string {
		if needsCreateVariant[name] || name == managementAPITenantConfigDef {
			return name + "CreateInput"
		}
		return name + "Input"
	}

	createInput := make(map[string]*jsonschema.Schema, len(needsCreateVariant)+1)
	buildCreateInput := func(name string) {
		s := cloneSchema(response[name])
		stripUnconditionalRequired(s, false)
		rewriteSchemaRefs(s, createInputResolve)
		createInput[name+"CreateInput"] = s
	}
	for name := range needsCreateVariant {
		buildCreateInput(name)
	}
	buildCreateInput(managementAPITenantConfigDef)

	var result []generatedSchema
	result = append(result,
		generatedSchema{managementAPITenantConfigDef, response[managementAPITenantConfigDef]},
		generatedSchema{managementAPITenantConfigDef + "Input", input[managementAPITenantConfigDef+"Input"]},
		generatedSchema{managementAPITenantConfigDef + "CreateInput", createInput[managementAPITenantConfigDef+"CreateInput"]},
	)
	result = append(result, sortedEntries(response, managementAPITenantConfigDef)...)
	result = append(result, sortedEntries(input, managementAPITenantConfigDef+"Input")...)
	result = append(result, sortedEntries(createInput, managementAPITenantConfigDef+"CreateInput")...)

	return result, nil
}

// sortedEntries returns m's entries as generatedSchemas, alphabetically by name, omitting exclude
// (which the caller already placed first).
func sortedEntries(m map[string]*jsonschema.Schema, exclude string) []generatedSchema {
	names := make([]string, 0, len(m))
	for name := range m {
		if name == exclude {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	entries := make([]generatedSchema, 0, len(names))
	for _, name := range names {
		entries = append(entries, generatedSchema{name, m[name]})
	}
	return entries
}

func identitySuffix(name string) string { return name }

func suffixWith(suffix string) func(string) string {
	return func(name string) string { return name + suffix }
}

// cloneSchema deep-copies a *jsonschema.Schema via a JSON round-trip — invopop's Schema type
// already marshals/unmarshals faithfully (including its ordered Properties map), so this is
// simpler and less error-prone than hand-writing a recursive copy.
func cloneSchema(s *jsonschema.Schema) *jsonschema.Schema {
	if s == nil {
		return nil
	}
	data, err := json.Marshal(s)
	if err != nil {
		panic(fmt.Errorf("failed to marshal schema for cloning: %w", err))
	}
	clone := new(jsonschema.Schema)
	if err := json.Unmarshal(data, clone); err != nil {
		panic(fmt.Errorf("failed to unmarshal schema for cloning: %w", err))
	}
	// invopop's MarshalJSON merges Extras (used for e.g. `meta:enum`) directly into the JSON
	// object, but UnmarshalJSON doesn't reverse that — unrecognized keys are just dropped. Restore
	// Extras explicitly at every level so re-cloning (response -> input -> createInput) doesn't
	// silently lose it.
	copyExtras(s, clone)
	return clone
}

// copyExtras copies .Extras from src to dst at every corresponding node, assuming dst was produced
// by cloning src (so their shapes match).
func copyExtras(src, dst *jsonschema.Schema) {
	if src == nil || dst == nil {
		return
	}

	dst.Extras = src.Extras

	if src.Properties != nil && dst.Properties != nil {
		for name, srcProp := range src.Properties.FromOldest() {
			if dstProp, ok := dst.Properties.Get(name); ok {
				copyExtras(srcProp, dstProp)
			}
		}
	}
	copyExtras(src.Items, dst.Items)
	copyExtras(src.AdditionalProperties, dst.AdditionalProperties)
	for key, srcProp := range src.PatternProperties {
		if dstProp, ok := dst.PatternProperties[key]; ok {
			copyExtras(srcProp, dstProp)
		}
	}
	copyExtras(src.If, dst.If)
	copyExtras(src.Then, dst.Then)
	copyExtras(src.Else, dst.Else)
	for i := range src.AllOf {
		if i < len(dst.AllOf) {
			copyExtras(src.AllOf[i], dst.AllOf[i])
		}
	}
	for i := range src.AnyOf {
		if i < len(dst.AnyOf) {
			copyExtras(src.AnyOf[i], dst.AnyOf[i])
		}
	}
	for i := range src.OneOf {
		if i < len(dst.OneOf) {
			copyExtras(src.OneOf[i], dst.OneOf[i])
		}
	}
}

func refName(s *jsonschema.Schema) string {
	if s == nil || s.Ref == "" {
		return ""
	}
	for i := len(s.Ref) - 1; i >= 0; i-- {
		if s.Ref[i] == '/' {
			return s.Ref[i+1:]
		}
	}
	return s.Ref
}

// computeRequired recursively forces every struct-shaped object's `required` to include all of its
// own declared properties, except optionalTenantConfigRefs.
func computeRequired(s *jsonschema.Schema) {
	if s == nil {
		return
	}

	if s.Properties != nil {
		required := make([]string, 0, s.Properties.Len())
		for name, prop := range s.Properties.FromOldest() {
			computeRequired(prop)
			if !optionalTenantConfigRefs[refName(prop)] {
				required = append(required, name)
			}
		}
		if s.Type == "object" {
			s.Required = required
		}
	}

	computeRequired(s.Items)
	computeRequired(s.AdditionalProperties)
	for _, prop := range s.PatternProperties {
		computeRequired(prop)
	}
}

// stripRequiredAndConditionals recursively removes `required`, `if`, `then`, and `else` everywhere
// — used to build the *Input (Update) variant, where none of those claims can be honestly made
// about a single partial request.
func stripRequiredAndConditionals(s *jsonschema.Schema) {
	if s == nil {
		return
	}

	s.Required = nil
	s.If = nil
	s.Then = nil
	s.Else = nil

	if s.Properties != nil {
		for _, prop := range s.Properties.FromOldest() {
			stripRequiredAndConditionals(prop)
		}
	}
	stripRequiredAndConditionals(s.Items)
	stripRequiredAndConditionals(s.AdditionalProperties)
	for _, prop := range s.PatternProperties {
		stripRequiredAndConditionals(prop)
	}
}

// stripUnconditionalRequired removes `required` outside of if/then/else subtrees, but preserves it
// inside them — used to build the *CreateInput variant, where conditional rules (the payload of
// then/else) stay meaningful but a flat "always required" claim doesn't.
func stripUnconditionalRequired(s *jsonschema.Schema, insideConditional bool) {
	if s == nil {
		return
	}

	if !insideConditional {
		s.Required = nil
	}

	stripUnconditionalRequired(s.If, true)
	stripUnconditionalRequired(s.Then, true)
	stripUnconditionalRequired(s.Else, true)

	if s.Properties != nil {
		for _, prop := range s.Properties.FromOldest() {
			stripUnconditionalRequired(prop, insideConditional)
		}
	}
	stripUnconditionalRequired(s.Items, insideConditional)
	stripUnconditionalRequired(s.AdditionalProperties, insideConditional)
	for _, prop := range s.PatternProperties {
		stripUnconditionalRequired(prop, insideConditional)
	}
}

// rewriteSchemaRefs rewrites every $ref in the tree to "#/components/schemas/<resolve(name)>".
func rewriteSchemaRefs(s *jsonschema.Schema, resolve func(name string) string) {
	if s == nil {
		return
	}

	if s.Ref != "" {
		s.Ref = "#/components/schemas/" + resolve(refName(s))
	}

	if s.Properties != nil {
		for _, prop := range s.Properties.FromOldest() {
			rewriteSchemaRefs(prop, resolve)
		}
	}
	rewriteSchemaRefs(s.Items, resolve)
	rewriteSchemaRefs(s.AdditionalProperties, resolve)
	for _, prop := range s.PatternProperties {
		rewriteSchemaRefs(prop, resolve)
	}
	for _, sub := range s.AllOf {
		rewriteSchemaRefs(sub, resolve)
	}
	for _, sub := range s.AnyOf {
		rewriteSchemaRefs(sub, resolve)
	}
	for _, sub := range s.OneOf {
		rewriteSchemaRefs(sub, resolve)
	}
	rewriteSchemaRefs(s.If, resolve)
	rewriteSchemaRefs(s.Then, resolve)
	rewriteSchemaRefs(s.Else, resolve)
}
