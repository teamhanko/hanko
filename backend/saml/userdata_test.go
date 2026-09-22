package saml

import (
	"testing"
	"time"

	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
	"github.com/stretchr/testify/assert"
	samlConfig "github.com/teamhanko/hanko/backend/v3/config"
)

func attributeValues(values ...string) types.Attribute {
	attributeValues := make([]types.AttributeValue, 0, len(values))
	for _, v := range values {
		attributeValues = append(attributeValues, types.AttributeValue{Value: v})
	}
	return types.Attribute{Values: attributeValues}
}

func TestBuildCustomClaimSource_NoMapping(t *testing.T) {
	source := buildCustomClaimSource(nil, saml2.Values{})

	assert.Nil(t, source)
}

func TestBuildCustomClaimSource_ResolvesMappedAttributes(t *testing.T) {
	values := saml2.Values{
		"urn:oid:mat_nr":      attributeValues("12345"),
		"urn:oid:affiliation": attributeValues("student", "staff"), // multi-valued
		"urn:oid:not-mapped":  attributeValues("irrelevant"),
	}
	mapping := map[string]string{
		"matriculation_number": "urn:oid:mat_nr",
		"affiliation":          "urn:oid:affiliation",
	}

	source := buildCustomClaimSource(mapping, values)

	assert.Equal(t, mapping, source.Mapping)
	assert.Equal(t, []string{"12345"}, source.Attributes["urn:oid:mat_nr"])
	assert.Equal(t, []string{"student", "staff"}, source.Attributes["urn:oid:affiliation"],
		"GetAll must be used, not Get, so multi-valued attributes aren't truncated to their first value")
	assert.NotContains(t, source.Attributes, "urn:oid:not-mapped")
}

func TestBuildCustomClaimSource_MissingAttributeOmitted(t *testing.T) {
	mapping := map[string]string{"matriculation_number": "urn:oid:mat_nr"}

	source := buildCustomClaimSource(mapping, saml2.Values{})

	assert.Equal(t, mapping, source.Mapping)
	assert.NotContains(t, source.Attributes, "urn:oid:mat_nr")
}

func TestExtractUserData_PopulatesCustomClaimSource(t *testing.T) {
	authnInstant := time.Now()
	assertionInfo := &saml2.AssertionInfo{
		AuthnInstant: &authnInstant,
		Values: saml2.Values{
			"email_attr":     attributeValues("alice@example.com"),
			"urn:oid:mat_nr": attributeValues("12345"),
		},
		Assertions: []types.Assertion{{
			Issuer:     &types.Issuer{Value: "https://idp.example.com"},
			Subject:    &types.Subject{NameID: &types.NameID{Value: "alice"}},
			Conditions: &types.Conditions{NotOnOrAfter: "2026-01-01T00:00:00Z"},
		}},
	}
	providerConfig := &ProviderConfig{
		AttributeMap: samlConfig.AttributeMap{
			Email:  "email_attr",
			Custom: map[string]string{"matriculation_number": "urn:oid:mat_nr"},
		},
	}

	userData := ExtractUserData(assertionInfo, providerConfig, "audience")

	assert.NotNil(t, userData.CustomClaimSource)
	assert.Equal(t, []string{"12345"}, userData.CustomClaimSource.Attributes["urn:oid:mat_nr"])
}
