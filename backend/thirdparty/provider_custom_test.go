package thirdparty

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCustomClaimSource_NoMapping(t *testing.T) {
	source := buildCustomClaimSource(nil, map[string]interface{}{"sub": "123"})

	assert.Nil(t, source)
}

func TestBuildCustomClaimSource_TopLevelClaim(t *testing.T) {
	rawClaims := map[string]interface{}{"groups": []interface{}{"staff", "student"}}
	mapping := map[string]string{"affiliation": "groups"}

	source := buildCustomClaimSource(mapping, rawClaims)

	assert.Equal(t, mapping, source.Mapping)
	assert.Equal(t, []interface{}{"staff", "student"}, source.Attributes["groups"])
}

func TestBuildCustomClaimSource_NestedGjsonPath(t *testing.T) {
	rawClaims := map[string]interface{}{
		"address": map[string]interface{}{"locality": "Hamburg"},
	}
	mapping := map[string]string{"city": "address.locality"}

	source := buildCustomClaimSource(mapping, rawClaims)

	assert.Equal(t, "Hamburg", source.Attributes["address.locality"])
}

func TestBuildCustomClaimSource_MissingPathOmitted(t *testing.T) {
	rawClaims := map[string]interface{}{"sub": "123"}
	mapping := map[string]string{"matriculation_number": "does.not.exist"}

	source := buildCustomClaimSource(mapping, rawClaims)

	assert.NotContains(t, source.Attributes, "does.not.exist")
}

func TestBuildCustomClaimSource_UnaffectedByLaterMutation(t *testing.T) {
	// Simulates the ordering hazard this function is built to avoid: GetUserData calls this
	// BEFORE AttributeMapping's rename/delete loop runs on rawClaims. Once this function has
	// returned, its result must not observe any later mutation of the same map.
	rawClaims := map[string]interface{}{"mat_nr": "12345"}
	mapping := map[string]string{"matriculation_number": "mat_nr"}

	source := buildCustomClaimSource(mapping, rawClaims)

	// Simulate AttributeMapping consuming/deleting the same key afterwards.
	delete(rawClaims, "mat_nr")

	assert.Equal(t, "12345", source.Attributes["mat_nr"],
		"must be unaffected by mutation of rawClaims after the snapshot was taken")
}
