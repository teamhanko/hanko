package thirdparty

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/v3/config"
)

func TestCustomClaimValueEqual(t *testing.T) {
	assert.True(t, customClaimValueEqual("student", "student"))
	assert.False(t, customClaimValueEqual("student", "staff"))
	assert.True(t, customClaimValueEqual(float64(29), float64(29)))
	assert.True(t, customClaimValueEqual(true, true))
	// []string (coerceCustomClaim's output) vs []interface{} (what a stored value round-trips
	// to via json.Unmarshal into an `any`) - different dynamic types, same content.
	assert.True(t, customClaimValueEqual([]string{"student", "staff"}, []interface{}{"student", "staff"}))
	assert.False(t, customClaimValueEqual([]string{"student"}, []interface{}{"student", "staff"}))
}

func TestMergeCustomClaims_FirstWriteAlwaysChanges(t *testing.T) {
	stored := map[string]StoredCustomClaim{}

	wrote, valueChanged := mergeCustomClaims(stored, map[string]any{"role": "student"}, []string{"role"}, "saml:uni-a")

	assert.True(t, wrote)
	assert.True(t, valueChanged)
	assert.Equal(t, StoredCustomClaim{Value: "student", Source: "saml:uni-a"}, stored["role"])
}

func TestMergeCustomClaims_SameConnectionSameValue_NoChange(t *testing.T) {
	stored := map[string]StoredCustomClaim{"role": {Value: "student", Source: "saml:uni-a"}}

	wrote, valueChanged := mergeCustomClaims(stored, map[string]any{"role": "student"}, []string{"role"}, "saml:uni-a")

	assert.False(t, wrote, "an IdP re-asserting an unchanged value must not produce a write")
	assert.False(t, valueChanged)
}

func TestMergeCustomClaims_SameConnectionDifferentValue_Changes(t *testing.T) {
	stored := map[string]StoredCustomClaim{"role": {Value: "student", Source: "saml:uni-a"}}

	wrote, valueChanged := mergeCustomClaims(stored, map[string]any{"role": "staff"}, []string{"role"}, "saml:uni-a")

	assert.True(t, wrote)
	assert.True(t, valueChanged)
	assert.Equal(t, "staff", stored["role"].Value)
}

func TestMergeCustomClaims_DifferentConnectionSameValue_WritesOwnershipButNotAChange(t *testing.T) {
	// uni-b now asserts the same value uni-a previously set - a real, anticipated scenario
	// (it's the reason "any source" write semantics exist at all). Ownership must transfer so
	// a later silence from uni-a can't incorrectly clear a claim uni-b is still maintaining -
	// but nothing looks different from outside, so no webhook should fire for it.
	stored := map[string]StoredCustomClaim{"role": {Value: "student", Source: "saml:uni-a"}}

	wrote, valueChanged := mergeCustomClaims(stored, map[string]any{"role": "student"}, []string{"role"}, "saml:uni-b")

	assert.True(t, wrote, "ownership handoff must still be persisted")
	assert.False(t, valueChanged, "same value from outside's perspective - must not fire a webhook")
	assert.Equal(t, StoredCustomClaim{Value: "student", Source: "saml:uni-b"}, stored["role"])
}

func TestMergeCustomClaims_OwnerSilence_ClearsClaim(t *testing.T) {
	stored := map[string]StoredCustomClaim{"role": {Value: "student", Source: "saml:uni-a"}}

	// uni-a still manages "role" (it's in `managed`) but no longer resolves a value for it.
	wrote, valueChanged := mergeCustomClaims(stored, map[string]any{}, []string{"role"}, "saml:uni-a")

	assert.True(t, wrote)
	assert.True(t, valueChanged)
	_, exists := stored["role"]
	assert.False(t, exists)
}

func TestMergeCustomClaims_NonOwnerSilence_LeavesClaimIntact(t *testing.T) {
	// After a handoff to uni-b (see the ownership-handoff test above), uni-a going quiet must
	// not be able to clear a claim it no longer owns.
	stored := map[string]StoredCustomClaim{"role": {Value: "student", Source: "saml:uni-b"}}

	wrote, valueChanged := mergeCustomClaims(stored, map[string]any{}, []string{"role"}, "saml:uni-a")

	assert.False(t, wrote)
	assert.False(t, valueChanged)
	assert.Equal(t, StoredCustomClaim{Value: "student", Source: "saml:uni-b"}, stored["role"])
}

func TestCustomClaimConnectionSource(t *testing.T) {
	assert.Equal(t, "saml:https://idp.example.com/metadata", customClaimConnectionSource("https://idp.example.com/metadata", true))
	assert.Equal(t, "third_party:custom_myprovider", customClaimConnectionSource("custom_myprovider", false))
}

func customClaimDefs() config.CustomClaimDefinitions {
	return config.CustomClaimDefinitions{
		"matriculation_number": {Name: "matriculation_number", Type: config.CustomClaimTypeString},
		"age":                  {Name: "age", Type: config.CustomClaimTypeNumber},
		"is_staff":             {Name: "is_staff", Type: config.CustomClaimTypeBoolean},
		"affiliation":          {Name: "affiliation", Type: config.CustomClaimTypeStringList},
	}
}

func TestResolveCustomClaims_NilSource(t *testing.T) {
	values, managed := ResolveCustomClaims(customClaimDefs(), nil)

	assert.Nil(t, values)
	assert.Nil(t, managed)
}

func TestResolveCustomClaims_AllFourTypes(t *testing.T) {
	src := &CustomClaimSource{
		Mapping: map[string]string{
			"matriculation_number": "mat_nr",
			"age":                  "age_attr",
			"is_staff":             "staff_attr",
			"affiliation":          "affiliation_attr",
		},
		Attributes: map[string]any{
			"mat_nr":           "12345",
			"age_attr":         "29",
			"staff_attr":       "true",
			"affiliation_attr": []string{"student", "staff"},
		},
	}

	values, managed := ResolveCustomClaims(customClaimDefs(), src)

	assert.Equal(t, "12345", values["matriculation_number"])
	assert.Equal(t, float64(29), values["age"])
	assert.Equal(t, true, values["is_staff"])
	assert.Equal(t, []string{"student", "staff"}, values["affiliation"])
	assert.ElementsMatch(t, []string{"matriculation_number", "age", "is_staff", "affiliation"}, managed)
}

func TestResolveCustomClaims_OIDCJSONShapes(t *testing.T) {
	src := &CustomClaimSource{
		Mapping: map[string]string{
			"age":         "age_claim",
			"is_staff":    "staff_claim",
			"affiliation": "affiliation_claim",
		},
		Attributes: map[string]any{
			"age_claim":         float64(29),
			"staff_claim":       true,
			"affiliation_claim": []interface{}{"student", "staff"},
		},
	}

	values, managed := ResolveCustomClaims(customClaimDefs(), src)

	assert.Equal(t, float64(29), values["age"])
	assert.Equal(t, true, values["is_staff"])
	assert.Equal(t, []string{"student", "staff"}, values["affiliation"])
	assert.ElementsMatch(t, []string{"age", "is_staff", "affiliation"}, managed)
}

func TestResolveCustomClaims_MissingAttribute(t *testing.T) {
	src := &CustomClaimSource{
		Mapping:    map[string]string{"matriculation_number": "mat_nr"},
		Attributes: map[string]any{},
	}

	values, managed := ResolveCustomClaims(customClaimDefs(), src)

	assert.Empty(t, values)
	assert.Equal(t, []string{"matriculation_number"}, managed, "still managed even though no value was found")
}

func TestResolveCustomClaims_UncoercibleValue(t *testing.T) {
	src := &CustomClaimSource{
		Mapping:    map[string]string{"age": "age_attr"},
		Attributes: map[string]any{"age_attr": "not-a-number"},
	}

	values, managed := ResolveCustomClaims(customClaimDefs(), src)

	assert.Empty(t, values, "coercion failure omits the claim, never fails the resolution")
	assert.Equal(t, []string{"age"}, managed)
}

func TestResolveCustomClaims_ObjectShapedValueRejected(t *testing.T) {
	src := &CustomClaimSource{
		Mapping: map[string]string{"matriculation_number": "address"},
		Attributes: map[string]any{
			"address": map[string]interface{}{"locality": "Hamburg"},
		},
	}

	values, managed := ResolveCustomClaims(customClaimDefs(), src)

	assert.Empty(t, values, "a JSON object must never be stringified into a claim value")
	assert.Equal(t, []string{"matriculation_number"}, managed)
}

func TestResolveCustomClaims_ListContainingObjectRejected(t *testing.T) {
	src := &CustomClaimSource{
		Mapping: map[string]string{"affiliation": "groups"},
		Attributes: map[string]any{
			"groups": []interface{}{"student", map[string]interface{}{"id": 1}},
		},
	}

	values, _ := ResolveCustomClaims(customClaimDefs(), src)

	assert.Empty(t, values)
}

func TestResolveCustomClaims_UndeclaredClaimSkippedNotManaged(t *testing.T) {
	src := &CustomClaimSource{
		Mapping:    map[string]string{"not_a_real_claim": "some_attr"},
		Attributes: map[string]any{"some_attr": "value"},
	}

	values, managed := ResolveCustomClaims(customClaimDefs(), src)

	assert.Empty(t, values)
	assert.Empty(t, managed, "an undeclared claim is neither resolved nor managed")
}

func TestResolveCustomClaims_BooleanGrammarMatchesParseBool(t *testing.T) {
	cases := map[string]bool{"true": true, "TRUE": true, "1": true, "t": true, "false": false, "0": false}
	for input, expected := range cases {
		src := &CustomClaimSource{
			Mapping:    map[string]string{"is_staff": "attr"},
			Attributes: map[string]any{"attr": input},
		}

		values, _ := ResolveCustomClaims(customClaimDefs(), src)

		assert.Equal(t, expected, values["is_staff"], "input %q", input)
	}

	// "yes" is deliberately not accepted - see decision 6 (align to strconv.ParseBool).
	src := &CustomClaimSource{
		Mapping:    map[string]string{"is_staff": "attr"},
		Attributes: map[string]any{"attr": "yes"},
	}
	values, _ := ResolveCustomClaims(customClaimDefs(), src)
	assert.Empty(t, values)
}
