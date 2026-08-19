package thirdparty

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/v3/config"
)

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
