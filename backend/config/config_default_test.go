package config

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DefaultTenantConfig() is the base every Create/Update fills gaps from — Create merges a partial
// request onto it, and Update merges onto whatever config a tenant already has, which itself
// originated from a Create call. If DefaultTenantConfig() were ever invalid, that invalidity would
// be silently inherited by every tenant that doesn't override the affected field, forever. This test
// pins down the one invariant everything else in the request-handling code relies on.
func TestDefaultTenantConfig_IsValid(t *testing.T) {
	tenantConfig := DefaultTenantConfig()

	err := tenantConfig.PostProcess()
	assert.NoError(t, err)

	err = tenantConfig.Validate(true)
	assert.NoError(t, err)
}

// schemaDefaultFormatExceptions lists paths where the schema's `default=` tag and
// DefaultTenantConfig()'s actual value legitimately disagree only in *representation*, not in
// meaning — not real drift, so excluded from the check below.
var schemaDefaultFormatExceptions = map[string]bool{
	// time.Duration has no custom MarshalJSON, so it always marshals as a plain number of
	// nanoseconds; the tag documents the YAML/file-config string convention ("720h") instead. Both
	// describe the same 720-hour duration.
	"mfa.device_trust_duration": true,
	// Both "0" and "0m" parse to a zero duration; just different spellings of "no idle timeout".
	"session.idle_timeout": true,
}

// TestDefaultTenantConfig_MatchesSchemaDefaults guards against the schema's documented `default=`
// annotations silently drifting from what config.DefaultTenantConfig() actually produces — the two
// are maintained independently (a struct tag vs. a Go literal) and nothing else keeps them in sync.
func TestDefaultTenantConfig_MatchesSchemaDefaults(t *testing.T) {
	reflector := &jsonschema.Reflector{}
	schema := reflector.Reflect(&TenantConfig{})

	raw, err := json.Marshal(DefaultTenantConfig())
	require.NoError(t, err)

	var actual interface{}
	require.NoError(t, json.Unmarshal(raw, &actual))

	var mismatches []string
	checkSchemaDefaults(schema, schema.Definitions, actual, "", &mismatches)
	assert.Empty(t, mismatches, "schema `default=` annotations that disagree with DefaultTenantConfig()")
}

func checkSchemaDefaults(schema *jsonschema.Schema, defs jsonschema.Definitions, actual interface{}, path string, mismatches *[]string) {
	if schema == nil {
		return
	}

	if schema.Ref != "" {
		name := refName(schema.Ref)
		resolved, ok := defs[name]
		if !ok {
			return
		}
		checkSchemaDefaults(resolved, defs, actual, path, mismatches)
		return
	}

	if schema.Default != nil && !schemaDefaultFormatExceptions[path] {
		wantJSON, err := json.Marshal(schema.Default)
		if err == nil {
			gotJSON, _ := json.Marshal(actual)
			if string(wantJSON) != string(gotJSON) {
				*mismatches = append(*mismatches, fmt.Sprintf("%s: schema default=%s, DefaultTenantConfig()=%s", path, wantJSON, gotJSON))
			}
		}
	}

	if schema.Properties != nil {
		actualMap, _ := actual.(map[string]interface{})
		for name, prop := range schema.Properties.FromOldest() {
			checkSchemaDefaults(prop, defs, actualMap[name], joinPath(path, name), mismatches)
		}
	}

	if schema.Items != nil {
		if actualSlice, ok := actual.([]interface{}); ok {
			for i, item := range actualSlice {
				checkSchemaDefaults(schema.Items, defs, item, fmt.Sprintf("%s[%d]", path, i), mismatches)
			}
		}
	}

	if schema.AdditionalProperties != nil {
		if actualMap, ok := actual.(map[string]interface{}); ok {
			for key, value := range actualMap {
				checkSchemaDefaults(schema.AdditionalProperties, defs, value, joinPath(path, key), mismatches)
			}
		}
	}
}

func refName(ref string) string {
	const prefix = "#/$defs/"
	if len(ref) > len(prefix) && ref[:len(prefix)] == prefix {
		return ref[len(prefix):]
	}
	return ref
}

func joinPath(path, name string) string {
	if path == "" {
		return name
	}
	return path + "." + name
}
